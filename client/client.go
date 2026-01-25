/*
   Snarky - Zero Knowledge Dead Drop
   Copyright (C) 2026 Sapadian LLC.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"snarky/crypto"
	"strings"
)

// Internal payload to hide filename inside encryption
type SecurePayload struct {
	Filename string `json:"filename"`
	Data     []byte `json:"data"` // Raw file bytes
}

// TrackedReader tracks progress for the progress bar
type TrackedReader struct {
	Reader io.Reader
	Total  int64
	Read   int64
}

func (tr *TrackedReader) Read(p []byte) (n int, err error) {
	n, err = tr.Reader.Read(p)
	tr.Read += int64(n)
	tr.printProgress()
	return
}

func (tr *TrackedReader) printProgress() {
	percent := float64(tr.Read) / float64(tr.Total) * 100
	barLen := 20
	filled := int(percent / 100 * float64(barLen))
	bar := strings.Repeat("=", filled) + strings.Repeat("-", barLen-filled)
	fmt.Printf("\r\033[36mProgress: [%s] %.2f%%\033[0m", bar, percent)
	if tr.Read == tr.Total {
		fmt.Println() // New line on completion
	}
}

func Send(serverURL, path string) {
	// 1. Validation
	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: File not found: %v\n", err)
		os.Exit(1)
	}

	// Hard Client-side limit (Soft check before server hard check)
	// We use 15MB here to account for JSON+Base64 overhead if server limit is ~10MB
	if fileInfo.Size() > 10*1024*1024 {
		fmt.Println("Error: File exceeds 10MB limit.")
		os.Exit(1)
	}

	fmt.Printf("Reading %s...\n", fileInfo.Name())
	fileData, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	// 2. Pack Payload (Metadata + Data)
	payload := SecurePayload{
		Filename: filepath.Base(path),
		Data:     fileData,
	}
	jsonBytes, _ := json.Marshal(payload)

	// 3. Encrypt
	fmt.Print("Encrypting... ")
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	// We encrypt the JSON structure
	encryptedStr, err := crypto.Encrypt(jsonBytes, key)
	if err != nil {
		panic(err)
	}
	fmt.Println("Done.")

	// 4. Upload with Progress
	bodyData := []byte(encryptedStr)
	reader := &TrackedReader{
		Reader: bytes.NewReader(bodyData),
		Total:  int64(len(bodyData)),
	}

	fmt.Println("Uploading to Dead Drop...")
	resp, err := http.Post(fmt.Sprintf("%s/upload", serverURL), "text/plain", reader)
	if err != nil {
		fmt.Printf("\nConnection error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("\nServer Error: %s\n", resp.Status)
		os.Exit(1)
	}

	uuidBytes, _ := io.ReadAll(resp.Body)
	uuid := string(uuidBytes)
	encodedKey := base64.URLEncoding.EncodeToString(key)

	// 5. Output Keys
	fmt.Println("\n[SECURE DROP CREATED]")
	fmt.Printf("ID:  %s\n", uuid)
	fmt.Printf("KEY: %s\n", encodedKey)
	fmt.Println("\nTo retrieve:")
	fmt.Printf("snarky get -id %s -key %s\n", uuid, encodedKey)
}

func Get(serverURL, id, keyStr string) {
	fmt.Println("Connecting to Dead Drop...")

	// 1. Download
	resp, err := http.Get(fmt.Sprintf("%s/download/%s", serverURL, id))
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Error: File not found, expired, or already retrieved.")
		os.Exit(1)
	}

	// Read content with progress
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		contentLength = 10 * 1024 * 1024 // Fallback estimate
	}

	reader := &TrackedReader{
		Reader: resp.Body,
		Total:  contentLength,
	}
	encryptedData, _ := io.ReadAll(reader)

	// 2. Decrypt
	fmt.Print("Decrypting... ")
	key, err := base64.URLEncoding.DecodeString(keyStr)
	if err != nil {
		fmt.Println("Error: Invalid key format.")
		os.Exit(1)
	}

	decryptedJson, err := crypto.Decrypt(string(encryptedData), key)
	if err != nil {
		fmt.Println("\nError: Decryption failed. Incorrect key?")
		os.Exit(1)
	}

	// 3. Unpack Payload
	var payload SecurePayload
	err = json.Unmarshal(decryptedJson, &payload)
	if err != nil {
		// Fallback for legacy text-only drops
		fmt.Println("\nWarning: Legacy text format detected.")
		os.Stdout.Write(decryptedJson)
		return
	}
	fmt.Println("Done.")

	// 4. Save to Disk
	outputName := "downloaded_" + payload.Filename
	err = os.WriteFile(outputName, payload.Data, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n[SUCCESS] Saved as: %s\n", outputName)
}

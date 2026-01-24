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
	"fmt"
	"io"
	"net/http"
	"os"
	"snarky/crypto"
)

func Send(serverURL, filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// 1. Generate Key locally
	key, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	// 2. Encrypt locally
	encryptedStr, err := crypto.Encrypt(data, key)
	if err != nil {
		panic(err)
	}

	// 3. Upload Blob
	resp, err := http.Post(fmt.Sprintf("%s/upload", serverURL), "text/plain", bytes.NewBufferString(encryptedStr))
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	uuidBytes, _ := io.ReadAll(resp.Body)
	uuid := string(uuidBytes)
	encodedKey := base64.URLEncoding.EncodeToString(key)

	// 4. Print Result
	fmt.Println("\n[SECURE DROP CREATED]")
	fmt.Printf("ID:  %s\n", uuid)
	fmt.Printf("KEY: %s\n", encodedKey)
	fmt.Println("\nTo retrieve:")
	fmt.Printf("snarky get -id %s -key %s\n", uuid, encodedKey)

	// Optional: URL format if you build a web frontend later
	// fmt.Printf("URL: %s/view#%s:%s\n", serverURL, uuid, encodedKey)
}

func Get(serverURL, id, keyStr string) {
	// 1. Download Blob
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

	encryptedData, _ := io.ReadAll(resp.Body)

	// 2. Decode Key
	key, err := base64.URLEncoding.DecodeString(keyStr)
	if err != nil {
		fmt.Println("Error: Invalid key format.")
		os.Exit(1)
	}

	// 3. Decrypt
	decryptedData, err := crypto.Decrypt(string(encryptedData), key)
	if err != nil {
		fmt.Println("Error: Decryption failed. Incorrect key?")
		os.Exit(1)
	}

	// 4. Output
	// We write to Stdout so users can redirect: snarky get ... > output.zip
	os.Stdout.Write(decryptedData)
}

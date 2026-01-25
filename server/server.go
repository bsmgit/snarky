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

package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Config holds server settings
type Config struct {
	Port        string `json:"port"`
	StoragePath string `json:"storage_path"`
	MaxFileSize int64  `json:"max_file_size"` // Bytes
	Retention   string `json:"retention"`     // e.g., "24h"
}

var config Config

func LoadConfig(path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		// Default config if file missing
		log.Println("Config file not found, using defaults.")
		config = Config{
			Port:        "8080",
			StoragePath: "./uploads",
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			Retention:   "24h",
		}
	} else {
		json.Unmarshal(file, &config)
	}

	// Ensure storage directory exists
	if _, err := os.Stat(config.StoragePath); os.IsNotExist(err) {
		os.MkdirAll(config.StoragePath, 0755)
	}
}

func Start(configPath string) {
	LoadConfig(configPath)

	// Parse retention duration
	retentionDuration, _ := time.ParseDuration(config.Retention)

	// Background cleanup routine (The Grim Reaper)
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			cleanupExpiredFiles(retentionDuration)
		}
	}()

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download/", handleDownload)

	addr := ":" + config.Port
	fmt.Printf("Snarky Daemon listening on %s (Storage: %s, MaxSize: %d bytes)\n", addr, config.StoragePath, config.MaxFileSize)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Enforce Configured File Size Limit
	r.Body = http.MaxBytesReader(w, r.Body, config.MaxFileSize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Upload rejected: too large")
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	// 2. Save Encrypted Blob to Disk
	id := uuid.New().String()
	filePath := filepath.Join(config.StoragePath, id)

	err = os.WriteFile(filePath, body, 0644)
	if err != nil {
		log.Printf("Disk write error: %v", err)
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return
	}

	log.Printf("Stored new item: %s (%d bytes)", id, len(body))
	w.Write([]byte(id))
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/download/"):]
	filePath := filepath.Join(config.StoragePath, id)

	// 1. Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found or already burnt", http.StatusNotFound)
		return
	}

	// 2. Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Read error", http.StatusInternalServerError)
		return
	}

	// 3. Burn (Delete) after reading
	// Note: In a production system, you might delay this slightly or use a specific "burn" flag
	os.Remove(filePath)
	log.Printf("Item retrieved and burned: %s", id)

	w.Write(data)
}

func cleanupExpiredFiles(ttl time.Duration) {
	files, err := os.ReadDir(config.StoragePath)
	if err != nil {
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err == nil {
			if time.Since(info.ModTime()) > ttl {
				os.Remove(filepath.Join(config.StoragePath, file.Name()))
				log.Printf("Expired and removed: %s", file.Name())
			}
		}
	}
}

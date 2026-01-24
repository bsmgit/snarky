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
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type StoreItem struct {
	EncryptedData string
	CreatedAt     time.Time
}

var (
	store = make(map[string]StoreItem)
	mutex = &sync.Mutex{}
)

func Start(port string) {
	// Background cleanup routine (TTL: 24 Hours)
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			mutex.Lock()
			count := 0
			for id, item := range store {
				if time.Since(item.CreatedAt) > 24*time.Hour {
					delete(store, id)
					count++
				}
			}
			if count > 0 {
				log.Printf("Cleaned up %d expired items", count)
			}
			mutex.Unlock()
		}
	}()

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/download/", handleDownload)

	addr := ":" + port
	fmt.Printf("Snarky Daemon listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 10MB to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	id := uuid.New().String()
	mutex.Lock()
	store[id] = StoreItem{EncryptedData: string(body), CreatedAt: time.Now()}
	mutex.Unlock()

	log.Printf("Stored new item: %s", id)
	w.Write([]byte(id))
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/download/"):]

	mutex.Lock()
	item, exists := store[id]
	if exists {
		delete(store, id) // Burn after reading
		log.Printf("Item retrieved and burned: %s", id)
	}
	mutex.Unlock()

	if !exists {
		http.Error(w, "File not found or already burnt", http.StatusNotFound)
		return
	}

	w.Write([]byte(item.EncryptedData))
}

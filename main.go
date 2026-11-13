package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// 1. Route utama (GET /)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := Response{
			Status:  "success",
			Message: "Halo! Backend Go berhasil berjalan di Railway.",
		}
		json.NewEncoder(w).Encode(res)
	})

	// 2. Route cek kesehatan server (GET /health)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := Response{
			Status:  "ok",
			Message: "Server sehat",
		}
		json.NewEncoder(w).Encode(res)
	})

	// 3. Ambil PORT dari Railway, jika kosong (di laptop) gunakan 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Server berjalan di port %s...\n", port)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
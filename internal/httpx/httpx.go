package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// JSON menulis respons JSON dengan status tertentu.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("gagal menulis respons: %v", err)
	}
}

// Error menulis format error standar: { "message": "..." }.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"message": message})
}

// ServerError mencatat error asli di log dan membalas 500 generik ke klien.
func ServerError(w http.ResponseWriter, err error) {
	log.Printf("ERROR: %v", err)
	Error(w, http.StatusInternalServerError, "Terjadi kesalahan pada server")
}

// DecodeJSON membaca body JSON (maks 1 MB) ke dst.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("body request kosong")
		}
		return errors.New("format JSON tidak valid")
	}
	return nil
}

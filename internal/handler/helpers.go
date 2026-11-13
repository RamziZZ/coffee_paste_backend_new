package handler

import (
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"unicode"

	"coffeepaste/internal/httpx"
)

const (
	maxRupiah   = 4_000_000_000 // batas kolom INT UNSIGNED
	maxImageURL = 60_000        // kolom image_url bertipe TEXT (64 KB)
)

// pathID membaca parameter {id} dari URL.
func pathID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id, err == nil && id > 0
}

// decodeOrFail membaca body JSON; bila gagal langsung membalas 422 dan mengembalikan false.
func decodeOrFail(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := httpx.DecodeJSON(w, r, dst); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "Data yang dikirim tidak valid: "+err.Error())
		return false
	}
	return true
}

func unprocessable(w http.ResponseWriter, msg string) {
	httpx.Error(w, http.StatusUnprocessableEntity, msg)
}

func respondNotFound(w http.ResponseWriter, msg string) {
	httpx.Error(w, http.StatusNotFound, msg)
}

func validEmail(s string) bool {
	if s == "" || len(s) > 150 {
		return false
	}
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s {
		return false
	}
	at := strings.LastIndex(s, "@")
	return strings.Contains(s[at+1:], ".")
}

// validPassword: minimal 8 karakter, ada huruf besar, huruf kecil, dan angka (sama dengan form frontend).
func validPassword(p string) bool {
	if len(p) < 8 || len(p) > 72 { // bcrypt hanya memakai 72 byte pertama
		return false
	}
	var up, low, digit bool
	for _, c := range p {
		switch {
		case unicode.IsUpper(c):
			up = true
		case unicode.IsLower(c):
			low = true
		case unicode.IsDigit(c):
			digit = true
		}
	}
	return up && low && digit
}

const passwordRuleMsg = "Password minimal 8 karakter (maksimal 72) dan harus mengandung huruf besar, huruf kecil, dan angka"

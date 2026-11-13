package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// validPhone menerima nomor seperti 0812-3456-7890 atau +62 812 3456 7890 (9-15 digit).
func validPhone(s string) bool {
	s = strings.NewReplacer(" ", "", "-", "").Replace(s)
	s = strings.TrimPrefix(s, "+")
	if len(s) < 9 || len(s) > 15 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Register: POST /api/auth/register  { name, email, phone, password }
// Membuat akun customer untuk aplikasi mobile. Tidak langsung login: aplikasi mengarahkan
// pengguna ke halaman login setelah berhasil (201).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		Password string `json:"password"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	switch {
	case name == "":
		unprocessable(w, "Nama wajib diisi")
		return
	case utf8.RuneCountInString(name) > 100:
		unprocessable(w, "Nama maksimal 100 karakter")
		return
	case !validEmail(email):
		unprocessable(w, "Format email tidak valid")
		return
	case phone != "" && !validPhone(phone):
		unprocessable(w, "Nomor telepon tidak valid")
		return
	case !validPassword(req.Password):
		unprocessable(w, passwordRuleMsg)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	user, err := h.users.CreateCustomer(r.Context(), name, email, phone, string(hash))
	if errors.Is(err, store.ErrDuplicateEmail) {
		httpx.Error(w, http.StatusConflict, "Email sudah terdaftar")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"message": "Akun berhasil dibuat", "user": user})
}

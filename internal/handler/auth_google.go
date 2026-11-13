package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"coffeepaste/internal/auth"
	"coffeepaste/internal/googleauth"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/model"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// WithGoogle mengaktifkan login Google. Tanpa ini, POST /api/auth/google membalas 503.
func (h *AuthHandler) WithGoogle(v googleauth.Verifier) *AuthHandler {
	h.google = v
	return h
}

// Google: POST /api/auth/google  { idToken }
// Aplikasi mobile mengirim ID token dari google_sign_in. Server memverifikasinya ke Google,
// lalu masuk / menghubungkan / membuat akun customer dan membalas { token, user }
// seperti /api/auth/login. Hanya akun customer yang boleh masuk lewat Google.
func (h *AuthHandler) Google(w http.ResponseWriter, r *http.Request) {
	if h.google == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "Login Google belum diaktifkan di server")
		return
	}

	var req struct {
		IDToken string `json:"idToken"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.IDToken) == "" {
		unprocessable(w, "idToken wajib diisi")
		return
	}

	claims, err := h.google.Verify(r.Context(), req.IDToken)
	if err != nil {
		log.Printf("login Google ditolak: %v", err)
		httpx.Error(w, http.StatusUnauthorized, "Token Google tidak valid atau sudah kedaluwarsa")
		return
	}
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if !claims.EmailVerified || !validEmail(email) {
		httpx.Error(w, http.StatusForbidden, "Email akun Google belum terverifikasi")
		return
	}

	user, err := h.resolveGoogleUser(r, claims, email)
	switch {
	case errors.Is(err, errGoogleNotCustomer):
		httpx.Error(w, http.StatusForbidden, "Akun ini bukan akun pelanggan. Gunakan email dan password untuk masuk.")
		return
	case errors.Is(err, store.ErrGoogleLinked):
		httpx.Error(w, http.StatusConflict, "Email ini sudah terhubung dengan akun Google lain")
		return
	case errors.Is(err, store.ErrDuplicateEmail):
		httpx.Error(w, http.StatusForbidden, "Akun dengan email ini tidak dapat dipakai untuk masuk")
		return
	case err != nil:
		httpx.ServerError(w, err)
		return
	}

	token, err := auth.NewToken(h.secret, h.ttl, user.ID, user.Role, user.TokenVersion)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

var errGoogleNotCustomer = errors.New("bukan akun customer")

// resolveGoogleUser: (1) sudah pernah login Google -> pakai akun itu; (2) email sudah terdaftar
// -> hubungkan; (3) selain itu -> buat akun customer baru.
func (h *AuthHandler) resolveGoogleUser(r *http.Request, c *googleauth.Claims, email string) (*model.User, error) {
	ctx := r.Context()

	user, err := h.users.FindByGoogleSub(ctx, c.Subject)
	if err == nil {
		if user.Role != "customer" {
			return nil, errGoogleNotCustomer
		}
		return user, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	existing, _, err := h.users.FindByEmail(ctx, email)
	if err == nil {
		if existing.Role != "customer" {
			return nil, errGoogleNotCustomer
		}
		if err := h.users.LinkGoogle(ctx, existing.ID, c.Subject); err != nil {
			return nil, err
		}
		return existing, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, err
	}

	name := strings.TrimSpace(c.Name)
	if name == "" {
		name, _, _ = strings.Cut(email, "@")
	}
	if utf8.RuneCountInString(name) > 100 {
		name = string([]rune(name)[:100])
	}

	// Kolom password_hash wajib terisi; isi hash acak yang tak pernah diketahui siapa pun.
	// Pengguna bisa membuat password sendiri lewat "Lupa Password".
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(hex.EncodeToString(raw)), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return h.users.CreateGoogleCustomer(ctx, name, email, c.Subject, string(hash))
}

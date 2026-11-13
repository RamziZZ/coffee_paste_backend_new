package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/mailer"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

const emailChangeTTLMinutes = 60

type AdminProfileHandler struct {
	users     *store.UserStore
	mail      mailer.Mailer
	publicURL string

	secret []byte        // untuk menerbitkan token baru setelah password diganti
	ttl    time.Duration
}

// WithTokens mengaktifkan penerbitan token baru setelah ganti password, supaya sesi
// perangkat ini tetap berlaku sementara sesi lain dicabut.
func (h *AdminProfileHandler) WithTokens(secret []byte, ttl time.Duration) *AdminProfileHandler {
	h.secret, h.ttl = secret, ttl
	return h
}

func NewAdminProfileHandler(users *store.UserStore, m mailer.Mailer, publicURL string) *AdminProfileHandler {
	return &AdminProfileHandler{users: users, mail: m, publicURL: publicURL}
}

func newToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Get: GET /api/admin/profile  -> { fullName, email, pendingEmail }
func (h *AdminProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	pending, err := h.users.PendingEmail(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"fullName":     user.Name,
		"email":        user.Email,
		"pendingEmail": pending,
	})
}

// UpdateName: PUT /api/admin/profile  { fullName }
func (h *AdminProfileHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		FullName string `json:"fullName"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.FullName)
	if name == "" {
		unprocessable(w, "Nama lengkap wajib diisi")
		return
	}
	if utf8.RuneCountInString(name) > 100 {
		unprocessable(w, "Nama lengkap maksimal 100 karakter")
		return
	}

	if err := h.users.UpdateName(r.Context(), user.ID, name); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"fullName": name})
}

// RequestEmailChange: POST /api/admin/profile/email  { email }
func (h *AdminProfileHandler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		Email           string `json:"email"`
		CurrentPassword string `json:"currentPassword"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	if !validEmail(email) {
		unprocessable(w, "Format email tidak valid")
		return
	}
	if !checkCurrentPassword(w, r, h.users, user.ID, req.CurrentPassword) {
		return
	}
	if strings.EqualFold(email, user.Email) {
		unprocessable(w, "Email baru sama dengan email saat ini")
		return
	}
	taken, err := h.users.EmailTaken(r.Context(), email, user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if taken {
		unprocessable(w, "Email sudah digunakan akun lain")
		return
	}

	token, hash, err := newToken()
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.CreateEmailChange(r.Context(), user.ID, email, hash, emailChangeTTLMinutes); err != nil {
		httpx.ServerError(w, err)
		return
	}

	link := fmt.Sprintf("%s/api/auth/email-change/verify?token=%s", h.publicURL, url.QueryEscape(token))
	body := "Halo,\n\n" +
		"Kami menerima permintaan mengganti email akun Coffee Paste Anda ke alamat ini.\n" +
		"Buka tautan berikut untuk mengonfirmasi (berlaku 60 menit):\n\n" + link + "\n\n" +
		"Jika Anda tidak merasa meminta ini, abaikan email ini.\n"

	if err := h.mail.Send(email, "Verifikasi email baru Coffee Paste", body); err != nil {
		log.Printf("gagal mengirim email verifikasi: %v", err)
		h.users.CancelEmailChange(r.Context(), user.ID)
		httpx.Error(w, http.StatusInternalServerError, "Gagal mengirim email verifikasi. Coba lagi nanti")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"pendingEmail": email})
}

// CancelEmailChange: DELETE /api/admin/profile/email
func (h *AdminProfileHandler) CancelEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if err := h.users.CancelEmailChange(r.Context(), user.ID); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ChangePassword: PUT /api/admin/profile/password  { newPassword }
func (h *AdminProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if !checkCurrentPassword(w, r, h.users, user.ID, req.CurrentPassword) {
		return
	}
	if !validPassword(req.NewPassword) {
		unprocessable(w, passwordRuleMsg)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.UpdatePassword(r.Context(), user.ID, string(hash)); err != nil {
		httpx.ServerError(w, err)
		return
	}
	respondPasswordChanged(w, r, h.users, h.secret, h.ttl, user)
}

// VerifyEmailChange: GET /api/auth/email-change/verify?token=...  (publik, dibuka dari tautan email)
// Membalas halaman HTML sederhana karena dibuka langsung di browser.
func (h *AdminProfileHandler) VerifyEmailChange(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))

	page := func(status int, title, msg string) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		fmt.Fprintf(w, `<!doctype html><html lang="id"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>%s</title>
<style>body{font-family:system-ui,sans-serif;display:flex;min-height:100vh;align-items:center;justify-content:center;margin:0;background:#faf7f2}
.card{background:#fff;padding:2rem 2.5rem;border-radius:12px;box-shadow:0 2px 12px rgba(0,0,0,.08);max-width:420px;text-align:center}
h1{font-size:1.25rem;margin:0 0 .75rem}p{color:#555;margin:0}</style></head>
<body><div class="card"><h1>%s</h1><p>%s</p></div></body></html>`, title, title, msg)
	}

	if token == "" {
		page(http.StatusBadRequest, "Tautan tidak valid", "Token verifikasi tidak ditemukan.")
		return
	}

	_, err := h.users.ConfirmEmailChange(r.Context(), hashToken(token))
	switch {
	case errors.Is(err, store.ErrNotFound):
		page(http.StatusBadRequest, "Tautan tidak berlaku", "Tautan sudah dipakai atau kedaluwarsa. Silakan ajukan ganti email lagi.")
	case errors.Is(err, store.ErrDuplicateEmail):
		page(http.StatusConflict, "Email sudah digunakan", "Email tersebut sudah dipakai akun lain.")
	case err != nil:
		log.Printf("ERROR: verifikasi email: %v", err)
		page(http.StatusInternalServerError, "Terjadi kesalahan", "Silakan coba lagi nanti.")
	default:
		page(http.StatusOK, "Email berhasil diganti", "Silakan masuk kembali memakai email baru Anda.")
	}
}

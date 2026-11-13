package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/mailer"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

const (
	resetTTLMinutes     = 60
	resetCooldownSecond = 60
)

type PasswordResetHandler struct {
	users       *store.UserStore
	mail        mailer.Mailer
	frontendURL string
}

func NewPasswordResetHandler(users *store.UserStore, m mailer.Mailer, frontendURL string) *PasswordResetHandler {
	return &PasswordResetHandler{users: users, mail: m, frontendURL: frontendURL}
}

// Forgot: POST /api/auth/forgot-password  { email }
// Selalu membalas 200 dengan pesan yang sama supaya tidak membocorkan email mana yang terdaftar.
// Metode: tautan berisi token acak (berlaku 60 menit, sekali pakai) dikirim ke email.
func (h *PasswordResetHandler) Forgot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	if !validEmail(email) {
		unprocessable(w, "Format email tidak valid")
		return
	}

	ok := func() {
		httpx.JSON(w, http.StatusOK, map[string]string{
			"message": "Jika email terdaftar, tautan reset password telah dikirim",
		})
	}

	user, _, err := h.users.FindByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) {
		ok()
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if recent, err := h.users.RecentResetExists(r.Context(), user.ID, resetCooldownSecond); err != nil {
		httpx.ServerError(w, err)
		return
	} else if recent {
		ok() // permintaan beruntun diabaikan diam-diam
		return
	}

	token, hash, err := newToken()
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.CreatePasswordReset(r.Context(), user.ID, hash, resetTTLMinutes); err != nil {
		httpx.ServerError(w, err)
		return
	}

	link := fmt.Sprintf("%s/reset-password?token=%s", h.frontendURL, url.QueryEscape(token))
	body := "Halo " + user.Name + ",\n\n" +
		"Kami menerima permintaan reset password akun Coffee Paste Anda.\n" +
		fmt.Sprintf("Buka tautan berikut untuk membuat password baru (berlaku %d menit):\n\n", resetTTLMinutes) +
		link + "\n\nJika Anda tidak merasa meminta ini, abaikan email ini.\n"
	if err := h.mail.Send(user.Email, "Reset password Coffee Paste", body); err != nil {
		log.Printf("gagal mengirim email reset password: %v", err)
		h.users.DeletePasswordReset(r.Context(), user.ID)
		httpx.Error(w, http.StatusInternalServerError, "Gagal mengirim email. Coba lagi nanti")
		return
	}
	ok()
}

// Reset: POST /api/auth/reset-password  { token, newPassword }
func (h *PasswordResetHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" {
		unprocessable(w, "Token reset wajib diisi")
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
	err = h.users.ResetPassword(r.Context(), hashToken(token), string(hash))
	if errors.Is(err, store.ErrNotFound) {
		unprocessable(w, "Tautan reset tidak valid atau sudah kedaluwarsa")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Password berhasil diubah. Silakan masuk dengan password baru"})
}

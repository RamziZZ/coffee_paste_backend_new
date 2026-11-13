package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/mailer"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

const (
	kasirCodeTTLMinutes = 15
	maxCodeAttempts     = 5 // salah memasukkan kode sebanyak ini membatalkan permintaan
)

type KasirProfileHandler struct {
	users *store.UserStore
	mail  mailer.Mailer

	mu       sync.Mutex
	failures map[int64]int // percobaan kode salah per user (di memori)

	secret []byte // untuk menerbitkan token baru setelah password diganti
	ttl    time.Duration
}

// WithTokens mengaktifkan penerbitan token baru setelah ganti password.
func (h *KasirProfileHandler) WithTokens(secret []byte, ttl time.Duration) *KasirProfileHandler {
	h.secret, h.ttl = secret, ttl
	return h
}

func NewKasirProfileHandler(users *store.UserStore, m mailer.Mailer) *KasirProfileHandler {
	return &KasirProfileHandler{users: users, mail: m, failures: make(map[int64]int)}
}

// codeHash mengikat kode 6 digit ke user dan email tujuan, sehingga kode tidak bisa
// dipakai untuk akun atau email lain.
func codeHash(userID int64, email, code string) string {
	return hashToken(fmt.Sprintf("kasir-email:%d:%s:%s", userID, strings.ToLower(email), code))
}

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (h *KasirProfileHandler) resetFailures(userID int64) {
	h.mu.Lock()
	delete(h.failures, userID)
	h.mu.Unlock()
}

// addFailure mencatat satu kode salah dan mengembalikan true bila batas sudah tercapai.
func (h *KasirProfileHandler) addFailure(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[userID]++
	if h.failures[userID] >= maxCodeAttempts {
		delete(h.failures, userID)
		return true
	}
	return false
}

// GetProfile: GET /api/kasir/profile  -> { fullName, email, phone, address }
func (h *KasirProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	phone, address, err := h.users.PhoneAddress(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{
		"fullName": user.Name,
		"email":    user.Email,
		"phone":    phone,
		"address":  address,
	})
}

// UpdateProfile: PUT /api/kasir/profile  { fullName, phone, address }
func (h *KasirProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		FullName string `json:"fullName"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.FullName)
	phone := strings.TrimSpace(req.Phone)
	address := strings.TrimSpace(req.Address)

	switch {
	case name == "":
		unprocessable(w, "Nama lengkap wajib diisi")
		return
	case utf8.RuneCountInString(name) > 100:
		unprocessable(w, "Nama lengkap maksimal 100 karakter")
		return
	case utf8.RuneCountInString(phone) > 30:
		unprocessable(w, "Nomor telepon maksimal 30 karakter")
		return
	case utf8.RuneCountInString(address) > 500:
		unprocessable(w, "Alamat maksimal 500 karakter")
		return
	}

	if err := h.users.UpdateProfile(r.Context(), user.ID, name, phone, address); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"fullName": name, "phone": phone, "address": address})
}

// RequestEmailChange: POST /api/kasir/profile/email  { email }  -> { pendingEmail }
// Kode 6 digit dikirim ke email baru.
func (h *KasirProfileHandler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

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

	code, err := newCode()
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.CreateEmailChange(r.Context(), user.ID, email, codeHash(user.ID, email, code), kasirCodeTTLMinutes); err != nil {
		httpx.ServerError(w, err)
		return
	}
	h.resetFailures(user.ID)

	body := "Halo,\n\n" +
		"Kode verifikasi untuk mengganti email akun Coffee Paste Anda: " + code + "\n" +
		fmt.Sprintf("Kode berlaku %d menit.\n\n", kasirCodeTTLMinutes) +
		"Jika Anda tidak merasa meminta ini, abaikan email ini.\n"
	if err := h.mail.Send(email, "Kode verifikasi email Coffee Paste", body); err != nil {
		log.Printf("gagal mengirim kode verifikasi: %v", err)
		h.users.CancelEmailChange(r.Context(), user.ID)
		httpx.Error(w, http.StatusInternalServerError, "Gagal mengirim kode verifikasi. Coba lagi nanti")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"pendingEmail": email})
}

// VerifyEmailChange: POST /api/kasir/profile/email/verify  { email, code }  -> { email }
func (h *KasirProfileHandler) VerifyEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)
	if len(code) != 6 || strings.Trim(code, "0123456789") != "" {
		unprocessable(w, "Kode verifikasi harus 6 digit angka")
		return
	}

	newEmail, err := h.users.ConfirmEmailChange(r.Context(), codeHash(user.ID, email, code))
	switch {
	case errors.Is(err, store.ErrNotFound):
		if h.addFailure(user.ID) {
			h.users.CancelEmailChange(r.Context(), user.ID)
			unprocessable(w, "Terlalu banyak kode salah. Ajukan ganti email lagi")
			return
		}
		unprocessable(w, "Kode verifikasi salah atau sudah kedaluwarsa")
	case errors.Is(err, store.ErrDuplicateEmail):
		unprocessable(w, "Email sudah digunakan akun lain")
	case err != nil:
		httpx.ServerError(w, err)
	default:
		h.resetFailures(user.ID)
		httpx.JSON(w, http.StatusOK, map[string]string{"email": newEmail})
	}
}

// ChangePassword: PUT /api/kasir/profile/password  { currentPassword, newPassword }
func (h *KasirProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if req.CurrentPassword == "" {
		unprocessable(w, "Password saat ini wajib diisi")
		return
	}
	if !validPassword(req.NewPassword) {
		unprocessable(w, passwordRuleMsg)
		return
	}

	hash, err := h.users.PasswordHash(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)) != nil {
		unprocessable(w, "Password saat ini salah")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.UpdatePassword(r.Context(), user.ID, string(newHash)); err != nil {
		httpx.ServerError(w, err)
		return
	}
	respondPasswordChanged(w, r, h.users, h.secret, h.ttl, user)
}

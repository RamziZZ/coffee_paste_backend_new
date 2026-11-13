package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/mailer"
	"coffeepaste/internal/model"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

const (
	otpTTLMinutes     = 10
	otpCooldownSecond = 60 // jeda minta ulang OTP; aplikasi mobile memakai angka yang sama untuk hitung mundur
)

// PasswordOTPHandler: lupa password dengan kode OTP 6 digit (aplikasi mobile).
// Sengaja terpisah dari PasswordResetHandler yang dipakai web (tautan token), jadi alur web tidak berubah.
type PasswordOTPHandler struct {
	users *store.UserStore
	mail  mailer.Mailer
}

func NewPasswordOTPHandler(users *store.UserStore, m mailer.Mailer) *PasswordOTPHandler {
	return &PasswordOTPHandler{users: users, mail: m}
}

// otpHash mengikat kode ke user, sehingga kode milik satu akun tidak berlaku untuk akun lain.
func otpHash(userID int64, code string) string {
	return hashToken(fmt.Sprintf("pwd-otp:%d:%s", userID, code))
}

func validOTPCode(code string) bool {
	return len(code) == 6 && strings.Trim(code, "0123456789") == ""
}

const msgOTPInvalid = "Kode OTP salah atau sudah kedaluwarsa"

// Request: POST /api/auth/forgot-password/otp  { email }
// Selalu membalas 200 dengan pesan yang sama supaya tidak membocorkan email mana yang terdaftar.
// Hanya akun customer yang menerima OTP (admin/kasir memakai web).
func (h *PasswordOTPHandler) Request(w http.ResponseWriter, r *http.Request) {
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
		httpx.JSON(w, http.StatusOK, map[string]any{
			"message":        "Jika email terdaftar, kode OTP telah dikirim",
			"expiresInMin":   otpTTLMinutes,
			"cooldownSecond": otpCooldownSecond,
		})
	}

	user, _, err := h.users.FindByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) || (err == nil && user.Role != "customer") {
		ok()
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if recent, err := h.users.RecentOTPExists(r.Context(), user.ID, otpCooldownSecond); err != nil {
		httpx.ServerError(w, err)
		return
	} else if recent {
		ok() // permintaan beruntun diabaikan diam-diam
		return
	}

	code, err := newCode()
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := h.users.CreatePasswordOTP(r.Context(), user.ID, otpHash(user.ID, code), otpTTLMinutes); err != nil {
		httpx.ServerError(w, err)
		return
	}

	body := "Halo " + user.Name + ",\n\n" +
		"Kode OTP untuk reset password akun Coffee Paste Anda: " + code + "\n" +
		fmt.Sprintf("Kode berlaku %d menit. Jangan bagikan kode ini kepada siapa pun.\n\n", otpTTLMinutes) +
		"Jika Anda tidak merasa meminta ini, abaikan email ini.\n"
	if err := h.mail.Send(user.Email, "Kode OTP reset password Coffee Paste", body); err != nil {
		log.Printf("gagal mengirim OTP reset password: %v", err)
		h.users.DeletePasswordOTP(r.Context(), user.ID)
		httpx.Error(w, http.StatusInternalServerError, "Gagal mengirim email. Coba lagi nanti")
		return
	}
	ok()
}

// Verify: POST /api/auth/forgot-password/otp/verify  { email, code }
// Hanya memeriksa kode (tidak memakainya) supaya aplikasi bisa lanjut ke layar password baru.
// Kode dicek ulang saat Reset, jadi langkah ini bukan jaminan keamanan tersendiri.
func (h *PasswordOTPHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)
	if !validEmail(email) {
		unprocessable(w, "Format email tidak valid")
		return
	}
	if !validOTPCode(code) {
		unprocessable(w, "Kode OTP harus 6 digit angka")
		return
	}

	user, ok := h.customerByEmail(w, r, email)
	if !ok {
		return
	}
	err := h.users.CheckPasswordOTP(r.Context(), user.ID, otpHash(user.ID, code), maxCodeAttempts)
	if h.otpFailed(w, err) {
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Kode OTP valid"})
}

// Reset: POST /api/auth/reset-password/otp  { email, code, newPassword }
func (h *PasswordOTPHandler) Reset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Code        string `json:"code"`
		NewPassword string `json:"newPassword"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	email := strings.TrimSpace(req.Email)
	code := strings.TrimSpace(req.Code)
	if !validEmail(email) {
		unprocessable(w, "Format email tidak valid")
		return
	}
	if !validOTPCode(code) {
		unprocessable(w, "Kode OTP harus 6 digit angka")
		return
	}
	if !validPassword(req.NewPassword) {
		unprocessable(w, passwordRuleMsg)
		return
	}

	user, ok := h.customerByEmail(w, r, email)
	if !ok {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	err = h.users.ResetPasswordWithOTP(r.Context(), user.ID, otpHash(user.ID, code), string(hash), maxCodeAttempts)
	if h.otpFailed(w, err) {
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Password berhasil diubah. Silakan masuk dengan password baru"})
}

// customerByEmail mencari akun customer aktif. Email tak dikenal / bukan customer dijawab
// persis seperti kode salah, supaya tidak membocorkan email yang terdaftar.
func (h *PasswordOTPHandler) customerByEmail(w http.ResponseWriter, r *http.Request, email string) (*model.User, bool) {
	user, _, err := h.users.FindByEmail(r.Context(), email)
	if errors.Is(err, store.ErrNotFound) || (err == nil && user.Role != "customer") {
		unprocessable(w, msgOTPInvalid)
		return nil, false
	}
	if err != nil {
		httpx.ServerError(w, err)
		return nil, false
	}
	return user, true
}

// otpFailed menulis respons error untuk hasil pemeriksaan OTP; true bila err bukan nil.
func (h *PasswordOTPHandler) otpFailed(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, store.ErrNotFound), errors.Is(err, store.ErrOTPInvalid):
		unprocessable(w, msgOTPInvalid)
	case errors.Is(err, store.ErrOTPExhausted):
		unprocessable(w, "Terlalu banyak kode salah. Minta kode OTP baru")
	default:
		httpx.ServerError(w, err)
	}
	return true
}

package handler

import (
	"errors"
	"fmt"
	"log"
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

// CustomerProfileHandler melayani profil dan ganti password untuk aplikasi mobile.
type CustomerProfileHandler struct {
	users  *store.UserStore
	secret []byte
	ttl    time.Duration

	mail     mailer.Mailer // pengirim kode verifikasi ganti email; nil = fitur ganti email nonaktif
	mu       sync.Mutex
	failures map[int64]int // percobaan kode salah per user (di memori)
}

func NewCustomerProfileHandler(users *store.UserStore, secret []byte, ttl time.Duration) *CustomerProfileHandler {
	return &CustomerProfileHandler{users: users, secret: secret, ttl: ttl, failures: make(map[int64]int)}
}

// WithMailer mengaktifkan ganti email (kode 6 digit dikirim ke email baru).
func (h *CustomerProfileHandler) WithMailer(m mailer.Mailer) *CustomerProfileHandler {
	h.mail = m
	return h
}

// Get: GET /api/profile -> { id, name, email, phone, gender, birthDate, pendingEmail }
func (h *CustomerProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	p, err := h.users.CustomerProfile(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if p.PendingEmail, err = h.users.PendingEmail(r.Context(), user.ID); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

// Update: PUT /api/profile  { name, phone, gender, birthDate }
// gender: "Laki-laki" | "Perempuan" | "" ; birthDate: "YYYY-MM-DD" | "". Email diubah lewat POST /api/profile/email.
func (h *CustomerProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		Name      string `json:"name"`
		Phone     string `json:"phone"`
		Gender    string `json:"gender"`
		BirthDate string `json:"birthDate"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)

	var gender, birth *string
	if req.Gender != "" {
		if req.Gender != "Laki-laki" && req.Gender != "Perempuan" {
			unprocessable(w, "Jenis kelamin harus Laki-laki atau Perempuan")
			return
		}
		g := req.Gender
		gender = &g
	}
	if req.BirthDate != "" {
		t, err := time.Parse("2006-01-02", req.BirthDate)
		if err != nil {
			unprocessable(w, "Tanggal lahir harus berformat YYYY-MM-DD")
			return
		}
		if t.After(time.Now()) {
			unprocessable(w, "Tanggal lahir tidak boleh di masa depan")
			return
		}
		b := req.BirthDate
		birth = &b
	}

	switch {
	case name == "":
		unprocessable(w, "Nama lengkap wajib diisi")
		return
	case utf8.RuneCountInString(name) > 100:
		unprocessable(w, "Nama lengkap maksimal 100 karakter")
		return
	case phone != "" && !validPhone(phone):
		unprocessable(w, "Nomor telepon tidak valid")
		return
	case utf8.RuneCountInString(phone) > 30:
		unprocessable(w, "Nomor telepon maksimal 30 karakter")
		return
	}

	if err := h.users.UpdateCustomerProfile(r.Context(), user.ID, name, phone, gender, birth); err != nil {
		httpx.ServerError(w, err)
		return
	}
	p, err := h.users.CustomerProfile(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

// ChangePassword: PUT /api/profile/password  { currentPassword, newPassword }
// Semua token lama dicabut, jadi respons berisi "token" baru yang harus disimpan aplikasi.
func (h *CustomerProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
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

func (h *CustomerProfileHandler) resetFailures(userID int64) {
	h.mu.Lock()
	delete(h.failures, userID)
	h.mu.Unlock()
}

// addFailure mencatat satu kode salah dan mengembalikan true bila batas sudah tercapai.
func (h *CustomerProfileHandler) addFailure(userID int64) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.failures[userID]++
	if h.failures[userID] >= maxCodeAttempts {
		delete(h.failures, userID)
		return true
	}
	return false
}

// RequestEmailChange: POST /api/profile/email  { email, currentPassword }  -> { pendingEmail }
// Kode 6 digit (berlaku 15 menit) dikirim ke email baru. Email lama tetap dipakai sampai kode diverifikasi.
func (h *CustomerProfileHandler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
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
	if strings.EqualFold(email, user.Email) {
		unprocessable(w, "Email baru sama dengan email saat ini")
		return
	}
	if h.mail == nil {
		httpx.Error(w, http.StatusServiceUnavailable, "Ganti email belum tersedia")
		return
	}
	if !checkCurrentPassword(w, r, h.users, user.ID, req.CurrentPassword) {
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

// VerifyEmailChange: POST /api/profile/email/verify  { email, code }  -> { email }
// Setelah berhasil, login berikutnya memakai email baru; token yang sedang dipakai tetap berlaku.
func (h *CustomerProfileHandler) VerifyEmailChange(w http.ResponseWriter, r *http.Request) {
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

// CancelEmailChange: DELETE /api/profile/email  -> membatalkan permintaan yang menunggu verifikasi.
func (h *CustomerProfileHandler) CancelEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	if err := h.users.CancelEmailChange(r.Context(), user.ID); err != nil {
		httpx.ServerError(w, err)
		return
	}
	h.resetFailures(user.ID)
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

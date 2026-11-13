package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"coffeepaste/internal/auth"
	"coffeepaste/internal/googleauth"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users     *store.UserStore
	secret    []byte
	ttl       time.Duration
	dummyHash []byte // dipakai agar waktu respons sama saat email tidak ditemukan
	google    googleauth.Verifier // nil = login Google nonaktif
}

func NewAuthHandler(users *store.UserStore, secret []byte, ttl time.Duration) *AuthHandler {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &AuthHandler{users: users, secret: secret, ttl: ttl, dummyHash: dummy}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // petunjuk dari form saja, tidak dipercaya server
}

// Login: POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, http.StatusUnprocessableEntity, "Email dan password wajib diisi")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		httpx.Error(w, http.StatusUnprocessableEntity, "Email dan password wajib diisi")
		return
	}

	const invalid = "Email atau password salah"

	user, hash, err := h.users.FindByEmail(r.Context(), req.Email)
	if errors.Is(err, store.ErrNotFound) {
		_ = bcrypt.CompareHashAndPassword(h.dummyHash, []byte(req.Password))
		httpx.Error(w, http.StatusUnauthorized, invalid)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		httpx.Error(w, http.StatusUnauthorized, invalid)
		return
	}

	token, err := auth.NewToken(h.secret, h.ttl, user.ID, user.Role, user.TokenVersion)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}

// Logout: POST /api/auth/logout
// Token JWT bersifat stateless, jadi server tidak menyimpan apa pun; frontend yang
// menghapus token. (Daftar token terblokir bisa ditambahkan nanti bila diperlukan.)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Berhasil keluar"})
}

// Me: GET /api/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
		return
	}
	httpx.JSON(w, http.StatusOK, user)
}

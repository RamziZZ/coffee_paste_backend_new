package handler

import (
	"net/http"
	"time"

	"coffeepaste/internal/auth"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/model"
	"coffeepaste/internal/store"

	"golang.org/x/crypto/bcrypt"
)

// checkCurrentPassword memastikan password saat ini benar. Bila tidak, langsung membalas
// 422 dan mengembalikan false.
func checkCurrentPassword(w http.ResponseWriter, r *http.Request, users *store.UserStore, userID int64, current string) bool {
	if current == "" {
		unprocessable(w, "Password saat ini wajib diisi")
		return false
	}
	hash, err := users.PasswordHash(r.Context(), userID)
	if err != nil {
		httpx.ServerError(w, err)
		return false
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(current)) != nil {
		unprocessable(w, "Password saat ini salah")
		return false
	}
	return true
}

// respondPasswordChanged membalas sukses ganti password. Karena token_version sudah naik,
// semua token lama (termasuk milik perangkat ini) tidak berlaku lagi, jadi token baru
// diterbitkan dan dikirim di field "token". Klien menyimpannya untuk menggantikan token lama.
func respondPasswordChanged(w http.ResponseWriter, r *http.Request, users *store.UserStore, secret []byte, ttl time.Duration, user *model.User) {
	resp := map[string]any{"ok": true}
	if len(secret) > 0 {
		if u, err := users.FindByID(r.Context(), user.ID); err == nil {
			if tok, err := auth.NewToken(secret, ttl, u.ID, u.Role, u.TokenVersion); err == nil {
				resp["token"] = tok
			}
		}
	}
	httpx.JSON(w, http.StatusOK, resp)
}

package middleware

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"coffeepaste/internal/auth"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/model"
	"coffeepaste/internal/store"
)

type userKey struct{}

// UserFromContext mengambil user yang sudah diautentikasi dari context request.
func UserFromContext(ctx context.Context) (*model.User, bool) {
	u, ok := ctx.Value(userKey{}).(*model.User)
	return u, ok
}

// Authenticate memeriksa header "Authorization: Bearer <token>" lalu memuat user dari database.
// User dimuat ulang tiap request supaya akun yang dinonaktifkan langsung kehilangan akses.
func Authenticate(secret []byte, users *store.UserStore) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			const prefix = "Bearer "
			header := r.Header.Get("Authorization")
			if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
				httpx.Error(w, http.StatusUnauthorized, "Token tidak ditemukan")
				return
			}

			claims, err := auth.ParseToken(secret, strings.TrimSpace(header[len(prefix):]))
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
				return
			}
			id, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
				return
			}

			user, err := users.FindByID(r.Context(), id)
			if errors.Is(err, store.ErrNotFound) {
				httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
				return
			}
			if err != nil {
				httpx.ServerError(w, err)
				return
			}

			// Password sudah diganti/direset sejak token ini diterbitkan -> sesi lama tidak berlaku.
			if claims.Ver != user.TokenVersion {
				httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
				return
			}

			ctx := context.WithValue(r.Context(), userKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole membatasi akses ke role tertentu. Harus dipasang setelah Authenticate.
func RequireRole(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
				return
			}
			if !slices.Contains(roles, user.Role) {
				httpx.Error(w, http.StatusForbidden, "Anda tidak memiliki akses ke fitur ini")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// TestContextWithUser menyisipkan user ke context. Hanya untuk pengujian handler.
func TestContextWithUser(ctx context.Context, u *model.User) context.Context {
	return context.WithValue(ctx, userKey{}, u)
}

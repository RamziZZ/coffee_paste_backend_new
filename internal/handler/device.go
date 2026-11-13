package handler

import (
	"net/http"
	"strings"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
)

// RemoveDevice: DELETE /api/devices  { token }
// Dipanggil saat logout agar perangkat tidak lagi menerima notifikasi akun tersebut.
func (h *MobileHandler) RemoveDevice(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	var req struct {
		Token string `json:"token"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	token := strings.TrimSpace(req.Token)
	if token == "" || len(token) > 255 {
		unprocessable(w, "Token perangkat wajib diisi (maksimal 255 karakter)")
		return
	}
	if err := h.users.RemoveDeviceToken(r.Context(), user.ID, token); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

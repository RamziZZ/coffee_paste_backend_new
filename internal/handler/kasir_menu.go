package handler

import (
	"errors"
	"net/http"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type KasirMenuHandler struct {
	menu *store.MenuStore
}

func NewKasirMenuHandler(menu *store.MenuStore) *KasirMenuHandler {
	return &KasirMenuHandler{menu: menu}
}

// List: GET /api/kasir/menu (bentuk data sama dengan /admin/menu)
func (h *KasirMenuHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.menu.ListAdmin(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

// SetStatus: PUT /api/kasir/menu/{id}/status  { status: "tersedia" | "habis" }
func (h *KasirMenuHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if req.Status != "tersedia" && req.Status != "habis" {
		unprocessable(w, "Status harus tersedia atau habis")
		return
	}

	item, err := h.menu.SetStatus(r.Context(), id, req.Status)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

package handler

import (
	"errors"
	"net/http"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type AdminReviewHandler struct {
	reviews *store.ReviewStore
}

func NewAdminReviewHandler(r *store.ReviewStore) *AdminReviewHandler {
	return &AdminReviewHandler{reviews: r}
}

// List: GET /api/admin/reviews
func (h *AdminReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.reviews.AdminList(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// SetPublished: PUT /api/admin/reviews/{id}  body { "isPublished": true|false }
func (h *AdminReviewHandler) SetPublished(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Review tidak ditemukan")
		return
	}
	var req struct {
		IsPublished *bool `json:"isPublished"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if req.IsPublished == nil {
		unprocessable(w, "Field isPublished wajib diisi")
		return
	}

	err := h.reviews.SetPublished(r.Context(), id, *req.IsPublished)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Review tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"id": id, "isPublished": *req.IsPublished})
}

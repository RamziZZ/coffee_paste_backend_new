package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type AdminMenuHandler struct {
	menu *store.MenuStore
}

func NewAdminMenuHandler(menu *store.MenuStore) *AdminMenuHandler {
	return &AdminMenuHandler{menu: menu}
}

type menuRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       int64   `json:"price"`      // harga Regular
	PriceLarge  *int64  `json:"priceLarge"` // harga Large (minuman); kosong = Regular + 4.000
	Status      string  `json:"status"`
	Image       *string `json:"image"`
	Description string  `json:"description"`
}

func (m menuRequest) toInput() (store.MenuInput, string) {
	name := strings.TrimSpace(m.Name)
	category := strings.TrimSpace(m.Category)
	desc := strings.TrimSpace(m.Description)

	switch {
	case name == "":
		return store.MenuInput{}, "Nama menu wajib diisi"
	case utf8.RuneCountInString(name) > 100:
		return store.MenuInput{}, "Nama menu maksimal 100 karakter"
	case category == "":
		return store.MenuInput{}, "Kategori wajib dipilih"
	case m.Price <= 0 || m.Price > maxRupiah:
		return store.MenuInput{}, "Harga harus lebih dari 0"
	case m.PriceLarge != nil && (*m.PriceLarge <= 0 || *m.PriceLarge > maxRupiah):
		return store.MenuInput{}, "Harga Large harus lebih dari 0"
	case m.Status != "tersedia" && m.Status != "habis":
		return store.MenuInput{}, "Status harus tersedia atau habis"
	case desc == "":
		return store.MenuInput{}, "Deskripsi wajib diisi"
	}

	// Foto dibuat opsional di server (frontend menampilkan placeholder bila kosong).
	var img *string
	if m.Image != nil {
		v := strings.TrimSpace(*m.Image)
		if len(v) > maxImageURL {
			return store.MenuInput{}, "Ukuran foto terlalu besar. Unggah lewat /admin/upload lalu kirim URL-nya"
		}
		if v != "" {
			img = &v
		}
	}

	return store.MenuInput{
		Name:        name,
		Category:    category,
		Price:       m.Price,
		PriceLarge:  m.PriceLarge,
		Status:      m.Status,
		Image:       img,
		Description: desc,
	}, ""
}

// Categories: GET /api/admin/menu/categories
func (h *AdminMenuHandler) Categories(w http.ResponseWriter, r *http.Request) {
	names, err := h.menu.CategoryNames(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, names)
}

// List: GET /api/admin/menu
func (h *AdminMenuHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.menu.ListAdmin(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

// Create: POST /api/admin/menu
func (h *AdminMenuHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req menuRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	item, err := h.menu.Create(r.Context(), in)
	if errors.Is(err, store.ErrInvalidCategory) {
		unprocessable(w, "Kategori tidak valid")
		return
	}
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, item)
}

// Update: PUT /api/admin/menu/{id}
func (h *AdminMenuHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	var req menuRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	item, err := h.menu.Update(r.Context(), id, in)
	if errors.Is(err, store.ErrInvalidCategory) {
		unprocessable(w, "Kategori tidak valid")
		return
	}
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
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

// Delete: DELETE /api/admin/menu/{id} (soft delete)
func (h *AdminMenuHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	err := h.menu.SoftDelete(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Menu berhasil dihapus"})
}

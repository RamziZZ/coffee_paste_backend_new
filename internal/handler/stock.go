package handler

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/store"
)

const (
	stockKasirHistoryDays = 7
	stockAdminHistoryDays = 90
)

var stockCategories = map[string]bool{
	"Bahan Baku": true,
	"Kemasan":    true,
	"Kebersihan": true,
	"Lainnya":    true,
}

// StockHandler melayani laporan bahan/stok habis: kasir melapor, admin memantau dan menandai sudah dibeli.
type StockHandler struct {
	stock *store.StockStore
}

func NewStockHandler(s *store.StockStore) *StockHandler {
	return &StockHandler{stock: s}
}

type stockReportRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Note     string `json:"note"`
}

func (q stockReportRequest) toInput() (store.StockInput, string) {
	name := strings.TrimSpace(q.Name)
	note := strings.TrimSpace(q.Note)

	switch {
	case name == "":
		return store.StockInput{}, "Nama bahan wajib diisi"
	case utf8.RuneCountInString(name) > 80:
		return store.StockInput{}, "Nama bahan maksimal 80 karakter"
	case !stockCategories[q.Category]:
		return store.StockInput{}, "Kategori harus Bahan Baku, Kemasan, Kebersihan, atau Lainnya"
	case utf8.RuneCountInString(note) > 200:
		return store.StockInput{}, "Catatan maksimal 200 karakter"
	}
	return store.StockInput{Name: name, Category: q.Category, Note: note}, ""
}

func stockError(w http.ResponseWriter, err error) {
	var ve store.ValidationError
	switch {
	case errors.As(err, &ve):
		unprocessable(w, ve.Error())
	case errors.Is(err, store.ErrNotFound):
		respondNotFound(w, "Laporan stok tidak ditemukan")
	default:
		httpx.ServerError(w, err)
	}
}

// KasirList: GET /api/kasir/stok-habis
func (h *StockHandler) KasirList(w http.ResponseWriter, r *http.Request) {
	list, err := h.stock.List(r.Context(), stockKasirHistoryDays)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// KasirCreate: POST /api/kasir/stok-habis  { name, category, note }
func (h *StockHandler) KasirCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
		return
	}
	var req stockReportRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	x, err := h.stock.Create(r.Context(), in, user.ID, user.Name)
	if err != nil {
		stockError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, x)
}

// KasirDelete: DELETE /api/kasir/stok-habis/{id}
func (h *StockHandler) KasirDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
		return
	}
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Laporan stok tidak ditemukan")
		return
	}
	if err := h.stock.DeleteOwnPending(r.Context(), id, user.ID); err != nil {
		stockError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Laporan stok berhasil dibatalkan"})
}

// AdminList: GET /api/admin/stok-habis
func (h *StockHandler) AdminList(w http.ResponseWriter, r *http.Request) {
	list, err := h.stock.List(r.Context(), stockAdminHistoryDays)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// AdminSetStatus: PUT /api/admin/stok-habis/{id}/status  { status: "habis" | "dibeli" }
func (h *StockHandler) AdminSetStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
		return
	}
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Laporan stok tidak ditemukan")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if req.Status != "habis" && req.Status != "dibeli" {
		unprocessable(w, "Status harus habis atau dibeli")
		return
	}

	x, err := h.stock.SetStatus(r.Context(), id, req.Status, user.ID)
	if err != nil {
		stockError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, x)
}

package handler

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/store"
)

type AdminExpenseHandler struct {
	expenses *store.ExpenseStore
}

func NewAdminExpenseHandler(e *store.ExpenseStore) *AdminExpenseHandler {
	return &AdminExpenseHandler{expenses: e}
}

type expenseRequest struct {
	Date        string `json:"date"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

func (e expenseRequest) toInput() (store.ExpenseInput, string) {
	name := strings.TrimSpace(e.Name)
	date := strings.TrimSpace(e.Date)

	if _, err := time.Parse("2006-01-02", date); err != nil {
		return store.ExpenseInput{}, "Tanggal harus berformat YYYY-MM-DD"
	}
	switch {
	case name == "":
		return store.ExpenseInput{}, "Nama pengeluaran wajib diisi"
	case utf8.RuneCountInString(name) > 150:
		return store.ExpenseInput{}, "Nama pengeluaran maksimal 150 karakter"
	case e.Category != "bahan-baku" && e.Category != "operasional" && e.Category != "lainnya":
		return store.ExpenseInput{}, "Kategori harus bahan-baku, operasional, atau lainnya"
	case e.Amount <= 0 || e.Amount > maxRupiah:
		return store.ExpenseInput{}, "Jumlah harus lebih dari 0"
	}

	return store.ExpenseInput{
		Date:        date,
		Name:        name,
		Category:    e.Category,
		Amount:      e.Amount,
		Description: strings.TrimSpace(e.Description),
	}, ""
}

// List: GET /api/admin/pengeluaran
func (h *AdminExpenseHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.expenses.List(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, list)
}

// Create: POST /api/admin/pengeluaran
func (h *AdminExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Sesi tidak valid atau sudah berakhir")
		return
	}
	var req expenseRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	x, err := h.expenses.Create(r.Context(), in, user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, x)
}

// Update: PUT /api/admin/pengeluaran/{id}
func (h *AdminExpenseHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Data pengeluaran tidak ditemukan")
		return
	}
	var req expenseRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	x, err := h.expenses.Update(r.Context(), id, in)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Data pengeluaran tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, x)
}

// Delete: DELETE /api/admin/pengeluaran/{id}
func (h *AdminExpenseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r)
	if !ok {
		respondNotFound(w, "Data pengeluaran tidak ditemukan")
		return
	}
	err := h.expenses.Delete(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Data pengeluaran tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "Pengeluaran berhasil dihapus"})
}

package handler

import (
	"net/http"
	"time"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type AdminOrderHandler struct {
	orders *store.OrderStore
}

func NewAdminOrderHandler(orders *store.OrderStore) *AdminOrderHandler {
	return &AdminOrderHandler{orders: orders}
}

func validDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// List: GET /api/admin/pesanan?from=YYYY-MM-DD&to=YYYY-MM-DD
// Filter tanggal opsional (inklusif). Tanpa filter, semua pesanan dikirim.
// Nomor telepon customer tidak disertakan (hanya untuk kasir).
func (h *AdminOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.OrderFilter{From: q.Get("from"), To: q.Get("to")}
	if (f.From != "" && !validDate(f.From)) || (f.To != "" && !validDate(f.To)) {
		unprocessable(w, "Format tanggal harus YYYY-MM-DD")
		return
	}
	if f.From != "" && f.To != "" && f.From > f.To {
		unprocessable(w, "Tanggal awal tidak boleh setelah tanggal akhir")
		return
	}

	orders, err := h.orders.List(r.Context(), f)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	for i := range orders {
		orders[i].Phone = nil
	}
	httpx.JSON(w, http.StatusOK, orders)
}

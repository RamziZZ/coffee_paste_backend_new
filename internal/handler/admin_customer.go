package handler

import (
	"net/http"

	"coffeepaste/internal/format"
	"coffeepaste/internal/httpx"
	"coffeepaste/internal/store"
)

type AdminCustomerHandler struct {
	customers *store.CustomerStore
}

func NewAdminCustomerHandler(c *store.CustomerStore) *AdminCustomerHandler {
	return &AdminCustomerHandler{customers: c}
}

type customerResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	TotalOrders int64  `json:"totalOrders"`
	TotalSpend  string `json:"totalSpend"` // string terformat "Rp 685.000", sesuai harapan frontend saat ini
	JoinedAt    string `json:"joinedAt"`
}

// List: GET /api/admin/customer
func (h *AdminCustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.customers.List(r.Context())
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	out := make([]customerResponse, 0, len(list))
	for _, c := range list {
		out = append(out, customerResponse{
			ID:          c.ID,
			Name:        c.Name,
			Email:       c.Email,
			Phone:       c.Phone,
			TotalOrders: c.TotalOrders,
			TotalSpend:  format.Rupiah(c.TotalSpend),
			JoinedAt:    c.JoinedAt.Format(format.ISO),
		})
	}
	httpx.JSON(w, http.StatusOK, out)
}

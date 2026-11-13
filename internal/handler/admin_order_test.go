package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminOrderListRejectsBadDates(t *testing.T) {
	h := NewAdminOrderHandler(nil) // validasi gagal sebelum menyentuh store
	for _, q := range []string{"from=2026-13-01", "to=besok", "from=2026-10-05&to=2026-10-01"} {
		w := httptest.NewRecorder()
		h.List(w, httptest.NewRequest("GET", "/api/admin/pesanan?"+q, nil))
		if w.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", q, w.Code)
		}
	}
}

func TestMobileValidation(t *testing.T) {
	h := NewMobileHandler(nil, nil, nil, nil)
	post := func(fn http.HandlerFunc, body string) int {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/", stringsReader(body))
		// handler membaca user dari context; tanpa user, panic dicegah lewat pengecekan di bawah
		fn(w, withUser(r))
		return w.Code
	}
	cases := []struct {
		name string
		fn   http.HandlerFunc
		body string
	}{
		{"rating 0", h.CreateReview, `{"rating":0,"comment":"enak"}`},
		{"rating 6", h.CreateReview, `{"rating":6,"comment":"enak"}`},
		{"komentar kosong", h.CreateReview, `{"rating":5,"comment":"  "}`},
		{"device platform salah", h.SaveDevice, `{"token":"abc","platform":"web"}`},
		{"device token kosong", h.SaveDevice, `{"token":"","platform":"ios"}`},
		{"favorit tanpa id", h.AddFavorite, `{}`},
		{"order tanpa menuItemId", h.CreateOrder, `{"orderType":"Take Away","paymentMethod":"QRIS","items":[{"name":"X","category":"Coffee","price":100,"qty":1}]}`},
	}
	for _, c := range cases {
		if code := post(c.fn, c.body); code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d, want 422", c.name, code)
		}
	}
}

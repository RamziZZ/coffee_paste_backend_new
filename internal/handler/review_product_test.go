package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"coffeepaste/internal/store"
)

// Semua kasus ditolak sebelum menyentuh database, jadi reviews boleh nil.
func TestReviewOrderValidation(t *testing.T) {
	h := NewMobileHandler(nil, nil, nil, nil)

	post := func(id, body string) int {
		req := httptest.NewRequest("POST", "/", stringsReader(body))
		req.SetPathValue("id", id)
		w := httptest.NewRecorder()
		h.ReviewOrder(w, withUser(req))
		return w.Code
	}

	cases := []struct {
		name, id, body string
		want           int
	}{
		{"id pesanan tidak valid", "abc", `{"reviews":[{"menuItemId":1,"rating":5}]}`, http.StatusNotFound},
		{"daftar kosong", "#CP00001", `{"reviews":[]}`, http.StatusUnprocessableEntity},
		{"tanpa field reviews", "1", `{}`, http.StatusUnprocessableEntity},
		{"rating 0", "1", `{"reviews":[{"menuItemId":1,"rating":0}]}`, http.StatusUnprocessableEntity},
		{"rating 6", "1", `{"reviews":[{"menuItemId":1,"rating":6}]}`, http.StatusUnprocessableEntity},
		{"menuItemId kosong", "1", `{"reviews":[{"rating":5}]}`, http.StatusUnprocessableEntity},
		{"komentar kepanjangan", "1", `{"reviews":[{"menuItemId":1,"rating":5,"comment":"` + strings.Repeat("a", 2001) + `"}]}`, http.StatusUnprocessableEntity},
		{"json rusak", "1", `{`, http.StatusUnprocessableEntity},
	}
	for _, c := range cases {
		if got := post(c.id, c.body); got != c.want {
			t.Errorf("%s: status %d, want %d", c.name, got, c.want)
		}
	}
}

func TestParseOrderCode(t *testing.T) {
	cases := map[string]int64{"#CP00012": 12, "CP00012": 12, "cp7": 7, "5": 5, " #CP1 ": 1}
	for in, want := range cases {
		if got, ok := parseOrderCode(in); !ok || got != want {
			t.Errorf("parseOrderCode(%q) = %d,%v; want %d", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "#", "CP", "0", "-3", "abc", "CPxyz"} {
		if _, ok := parseOrderCode(bad); ok {
			t.Errorf("parseOrderCode(%q) seharusnya tidak valid", bad)
		}
	}
}

func TestRound1(t *testing.T) {
	if got := round1(4.449); got != 4.4 {
		t.Errorf("round1(4.449) = %v", got)
	}
	if got := round1(4.45); got != 4.5 {
		t.Errorf("round1(4.45) = %v", got)
	}
	if got := round1(0); got != 0 {
		t.Errorf("round1(0) = %v", got)
	}
}

func TestToMobileOrderReviewedDefaultsToEmptySlice(t *testing.T) {
	m := toMobileOrder(storeOrderForTest())
	if m.ReviewedItemIDs == nil {
		t.Error("ReviewedItemIDs harus [] (bukan null) supaya JSON-nya konsisten")
	}
}

func storeOrderForTest() store.Order { return store.Order{} }

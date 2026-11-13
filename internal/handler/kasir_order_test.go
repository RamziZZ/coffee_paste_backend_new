package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func TestOrderPathID(t *testing.T) {
	cases := map[string]struct {
		id int64
		ok bool
	}{
		"#CP00012": {12, true},
		"CP00012":  {12, true},
		"cp00012":  {12, true},
		"12":       {12, true},
		"":         {0, false},
		"abc":      {0, false},
		"#CP":      {0, false},
		"0":        {0, false},
		"-3":       {0, false},
	}
	for in, want := range cases {
		mux := http.NewServeMux()
		var gotID int64
		var gotOK bool
		mux.HandleFunc("GET /x/{id}", func(w http.ResponseWriter, r *http.Request) {
			gotID, gotOK = orderPathID(r)
		})
		if in == "" {
			continue // segmen kosong tidak cocok dengan pola rute
		}
		req := httptest.NewRequest("GET", "/x/"+strings.ReplaceAll(in, "#", "%23"), nil)
		mux.ServeHTTP(httptest.NewRecorder(), req)
		if gotOK != want.ok || (want.ok && gotID != want.id) {
			t.Errorf("orderPathID(%q) = %d,%v; want %d,%v", in, gotID, gotOK, want.id, want.ok)
		}
	}
}

func validReq() createOrderRequest {
	return createOrderRequest{
		OrderType:     "Dine In",
		TableNumber:   ptr("Meja 4"),
		PaymentMethod: "QRIS",
		Items:         []orderItemRequest{{MenuItemID: ptr(int64(1)), Qty: 2}},
	}
}

func TestCreateOrderValidation(t *testing.T) {
	if _, msg := validReq().toInput(); msg != "" {
		t.Fatalf("request valid ditolak: %s", msg)
	}

	bad := map[string]func(*createOrderRequest){
		"tipe salah":            func(q *createOrderRequest) { q.OrderType = "Delivery" },
		"metode salah":          func(q *createOrderRequest) { q.PaymentMethod = "Kartu" },
		"tanpa item":            func(q *createOrderRequest) { q.Items = nil },
		"qty nol":               func(q *createOrderRequest) { q.Items[0].Qty = 0 },
		"qty terlalu besar":     func(q *createOrderRequest) { q.Items[0].Qty = 1000 },
		"ukuran salah":          func(q *createOrderRequest) { q.Items[0].Size = ptr("XL") },
		"es salah":             func(q *createOrderRequest) { q.Items[0].Ice = ptr("Banyak") },
		"gula salah":           func(q *createOrderRequest) { q.Items[0].Sugar = ptr("Manis") },
		"meja untuk take away":  func(q *createOrderRequest) { q.OrderType = "Take Away" },
		"item bebas tanpa nama": func(q *createOrderRequest) { q.Items[0] = orderItemRequest{Category: "Coffee", Price: 100, Qty: 1} },
		"item bebas harga nol":  func(q *createOrderRequest) { q.Items[0] = orderItemRequest{Name: "X", Category: "Coffee", Qty: 1} },
		"menuItemId negatif":    func(q *createOrderRequest) { q.Items[0].MenuItemID = ptr(int64(-1)) },
		"uang negatif":          func(q *createOrderRequest) { q.CashReceived = ptr(int64(-5)) },
	}
	for name, mutate := range bad {
		q := validReq()
		q.Items = append([]orderItemRequest(nil), q.Items...)
		mutate(&q)
		if _, msg := q.toInput(); msg == "" {
			t.Errorf("%s: seharusnya ditolak", name)
		}
	}

	// Take Away dengan nomor meja kosong/null boleh.
	q := validReq()
	q.OrderType = "Take Away"
	q.TableNumber = ptr("  ")
	in, msg := q.toInput()
	if msg != "" || in.TableNumber != nil {
		t.Errorf("take away dengan meja kosong: msg=%q table=%v", msg, in.TableNumber)
	}
}

func TestCodeHashBoundToUserAndEmail(t *testing.T) {
	a := codeHash(1, "a@x.com", "123456")
	if a != codeHash(1, "A@X.com", "123456") {
		t.Error("email harus tidak peka huruf besar/kecil")
	}
	if a == codeHash(2, "a@x.com", "123456") {
		t.Error("hash harus berbeda antar user")
	}
	if a == codeHash(1, "b@x.com", "123456") {
		t.Error("hash harus berbeda antar email")
	}
	if a == codeHash(1, "a@x.com", "654321") {
		t.Error("hash harus berbeda antar kode")
	}
}

func TestNewCodeFormat(t *testing.T) {
	for i := 0; i < 200; i++ {
		c, err := newCode()
		if err != nil || len(c) != 6 || strings.Trim(c, "0123456789") != "" {
			t.Fatalf("kode tidak valid: %q %v", c, err)
		}
	}
}

func TestEmailChangeAttemptLimit(t *testing.T) {
	h := NewKasirProfileHandler(nil, nil)
	for i := 1; i < maxCodeAttempts; i++ {
		if h.addFailure(7) {
			t.Fatalf("batas tercapai terlalu cepat pada percobaan %d", i)
		}
	}
	if !h.addFailure(7) {
		t.Error("batas seharusnya tercapai")
	}
	if h.addFailure(7) {
		t.Error("penghitung harus di-reset setelah batas tercapai")
	}
}

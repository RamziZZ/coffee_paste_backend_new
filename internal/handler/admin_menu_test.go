package handler

import "testing"

func TestMenuRequestPriceLarge(t *testing.T) {
	base := func() menuRequest {
		return menuRequest{Name: "Latte", Category: "Coffee", Price: 22000, Status: "tersedia", Description: "enak"}
	}

	// Tanpa priceLarge: diteruskan sebagai nil (server mengisi Regular + 4.000 untuk minuman).
	if in, msg := base().toInput(); msg != "" || in.PriceLarge != nil {
		t.Fatalf("tanpa priceLarge: msg=%q priceLarge=%v", msg, in.PriceLarge)
	}

	// Dengan priceLarge: diteruskan apa adanya.
	r := base()
	r.PriceLarge = ptr(int64(27000))
	if in, msg := r.toInput(); msg != "" || in.PriceLarge == nil || *in.PriceLarge != 27000 {
		t.Fatalf("dengan priceLarge: msg=%q priceLarge=%v", msg, in.PriceLarge)
	}

	// priceLarge nol, negatif, atau melebihi batas ditolak.
	for _, bad := range []int64{0, -1, maxRupiah + 1} {
		r := base()
		r.PriceLarge = ptr(bad)
		if _, msg := r.toInput(); msg == "" {
			t.Errorf("priceLarge %d seharusnya ditolak", bad)
		}
	}
}

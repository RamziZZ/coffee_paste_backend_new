package store

import (
	"strconv"
	"testing"
)

func TestUnitPrice(t *testing.T) {
	large := int64(26000)
	regular, lg := "Regular", "Large"

	cases := []struct {
		name  string
		large *int64
		size  *string
		want  int64
	}{
		{"regular memakai harga Regular", &large, &regular, 22000},
		{"tanpa ukuran memakai harga Regular", &large, nil, 22000},
		{"large memakai harga Large", &large, &lg, 26000},
		{"large tanpa harga Large = Regular + selisih bawaan", nil, &lg, 26000},
	}
	for _, c := range cases {
		if got := UnitPrice(22000, c.large, c.size); got != c.want {
			t.Errorf("%s: UnitPrice = %d, want %d", c.name, got, c.want)
		}
	}

	custom := int64(30000)
	if got := UnitPrice(22000, &custom, &lg); got != 30000 {
		t.Errorf("harga Large kustom: got %d, want 30000", got)
	}
}

func TestDefaultSurchargeConstantsAgree(t *testing.T) {
	if defaultLargeSurchargeSQL != strconv.FormatInt(DefaultLargeSurcharge, 10) {
		t.Fatalf("defaultLargeSurchargeSQL (%s) harus sama dengan DefaultLargeSurcharge (%d)",
			defaultLargeSurchargeSQL, DefaultLargeSurcharge)
	}
}

func TestLargeForSave(t *testing.T) {
	v := func(n int64) *int64 { return &n }

	// Non-minuman: selalu NULL, apa pun inputnya.
	if got, err := largeForSave(MenuInput{Price: 28000, PriceLarge: v(99000)}, false); err != nil || got != nil {
		t.Errorf("non-minuman: got %v, %v; want nil, nil", got, err)
	}
	// Minuman tanpa input: Regular + 4.000.
	if got, err := largeForSave(MenuInput{Price: 22000}, true); err != nil || got == nil || *got != 26000 {
		t.Errorf("default: got %v, %v; want 26000", got, err)
	}
	// Minuman dengan input kustom dipakai apa adanya.
	if got, err := largeForSave(MenuInput{Price: 22000, PriceLarge: v(25000)}, true); err != nil || got == nil || *got != 25000 {
		t.Errorf("kustom: got %v, %v; want 25000", got, err)
	}
	// Large sama dengan Regular masih boleh.
	if got, err := largeForSave(MenuInput{Price: 22000, PriceLarge: v(22000)}, true); err != nil || got == nil || *got != 22000 {
		t.Errorf("sama: got %v, %v; want 22000", got, err)
	}
	// Large lebih murah dari Regular ditolak.
	if _, err := largeForSave(MenuInput{Price: 22000, PriceLarge: v(21000)}, true); err == nil {
		t.Error("Large < Regular harus ditolak")
	}
}

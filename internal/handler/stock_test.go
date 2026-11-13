package handler

import (
	"strings"
	"testing"
)

func TestStockReportValidation(t *testing.T) {
	ok := stockReportRequest{Name: "  Susu  ", Category: "Bahan Baku", Note: " sisa 1 dus "}
	in, msg := ok.toInput()
	if msg != "" {
		t.Fatalf("request valid ditolak: %s", msg)
	}
	if in.Name != "Susu" || in.Note != "sisa 1 dus" {
		t.Errorf("input tidak dirapikan: %+v", in)
	}

	bad := map[string]stockReportRequest{
		"nama kosong":          {Name: "   ", Category: "Bahan Baku"},
		"nama terlalu panjang": {Name: strings.Repeat("a", 81), Category: "Bahan Baku"},
		"kategori salah":       {Name: "Susu", Category: "Minuman"},
		"catatan panjang":      {Name: "Susu", Category: "Kemasan", Note: strings.Repeat("a", 201)},
	}
	for name, req := range bad {
		if _, msg := req.toInput(); msg == "" {
			t.Errorf("%s: seharusnya ditolak", name)
		}
	}
}

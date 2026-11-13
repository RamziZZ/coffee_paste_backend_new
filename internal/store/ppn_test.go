package store

import "testing"

func TestCalcPPN(t *testing.T) {
	cases := map[int64]int64{
		0:      0,
		1000:   110,
		7000:   770,
		20000:  2200,
		36000:  3960,
		24000:  2640,
		4:      0,
		5:      1,
		50:     6,
		12_345: 1358,
	}
	for subtotal, want := range cases {
		if got := CalcPPN(subtotal); got != want {
			t.Errorf("CalcPPN(%d) = %d, want %d", subtotal, got, want)
		}
	}
}

func TestMaxSubtotalFitsUnsignedInt(t *testing.T) {
	if got := int64(maxSubtotal) + CalcPPN(maxSubtotal); got > 4_294_967_295 {
		t.Fatalf("subtotal maksimum + PPN = %d melebihi INT UNSIGNED", got)
	}
}

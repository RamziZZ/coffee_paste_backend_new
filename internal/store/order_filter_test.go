package store

import "testing"

func TestOrderFilterWhere(t *testing.T) {
	w, a := OrderFilter{}.where()
	if w != "" || len(a) != 0 {
		t.Errorf("filter kosong: %q %v", w, a)
	}
	w, a = OrderFilter{From: "2026-10-01", To: "2026-10-03", CustomerID: 5}.where()
	want := " WHERE o.created_at >= ? AND o.created_at < DATE_ADD(?, INTERVAL 1 DAY) AND o.customer_id = ?"
	if w != want || len(a) != 3 {
		t.Errorf("where=%q args=%v", w, a)
	}
	w, a = OrderFilter{To: "2026-10-03"}.where()
	if len(a) != 1 || a[0] != "2026-10-03" {
		t.Errorf("hanya To: %q %v", w, a)
	}
}

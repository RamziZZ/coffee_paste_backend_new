package store

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

// Tes integrasi harga Regular/Large: butuh MySQL/MariaDB dengan skema proyek.
// Jalankan: MENU_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/db?parseTime=true&clientFoundRows=true' go test ./internal/store -run MenuPrice
func TestMenuPriceFlow(t *testing.T) {
	dsn := os.Getenv("MENU_TEST_DSN")
	if dsn == "" {
		t.Skip("MENU_TEST_DSN belum diisi")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx := context.Background()

	menu := &MenuStore{DB: db}
	orders := &OrderStore{DB: db}
	for i := 0; i < 2; i++ { // idempoten
		if err := menu.EnsurePriceSchema(ctx); err != nil {
			t.Fatal(err)
		}
	}

	v := func(n int64) *int64 { return &n }
	drink, err := menu.Create(ctx, MenuInput{Name: "Tes Latte", Category: "Coffee", Price: 20000, Status: "tersedia", Description: "x"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { menu.SoftDelete(ctx, drink.ID) })
	if drink.PriceLarge == nil || *drink.PriceLarge != 24000 {
		t.Fatalf("default Large harus 24000, got %v", drink.PriceLarge)
	}

	custom, err := menu.Create(ctx, MenuInput{Name: "Tes Matcha", Category: "Non-Coffee", Price: 20000, PriceLarge: v(27500), Status: "tersedia", Description: "x"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { menu.SoftDelete(ctx, custom.ID) })
	if custom.PriceLarge == nil || *custom.PriceLarge != 27500 {
		t.Fatalf("Large kustom harus 27500, got %v", custom.PriceLarge)
	}

	food, err := menu.Create(ctx, MenuInput{Name: "Tes Roti", Category: "Pastry", Price: 15000, PriceLarge: v(99999), Status: "tersedia", Description: "x"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { menu.SoftDelete(ctx, food.ID) })
	if food.PriceLarge != nil {
		t.Fatalf("makanan tidak boleh punya harga Large, got %v", *food.PriceLarge)
	}

	if _, err := menu.Create(ctx, MenuInput{Name: "Tes Salah", Category: "Coffee", Price: 20000, PriceLarge: v(1000), Status: "tersedia", Description: "x"}); err == nil {
		t.Fatal("Large lebih murah dari Regular harus ditolak")
	}

	// Pesanan: harga satuan mengikuti ukuran, dihitung server dari database.
	size := func(s string) *string { return &s }
	order, err := orders.Create(ctx, NewOrder{
		OrderType: "Dine In", PaymentMethod: "QRIS",
		Items: []NewOrderItem{
			{MenuItemID: &drink.ID, Qty: 1, Size: size("Regular"), Price: 1}, // harga klien diabaikan
			{MenuItemID: &drink.ID, Qty: 2, Size: size("Large"), Price: 1},
			{MenuItemID: &custom.ID, Qty: 1, Size: size("Large")},
			{MenuItemID: &food.ID, Qty: 1, Size: size("Large")}, // ukuran diabaikan untuk makanan
		},
	}, 2)
	if err != nil {
		t.Fatal(err)
	}
	orderID, err := strconv.ParseInt(strings.TrimLeft(order.ID, "#CPcp"), 10, 64) // "#CP00012" -> 12
	if err != nil {
		t.Fatalf("kode pesanan %q tidak terbaca: %v", order.ID, err)
	}
	t.Cleanup(func() {
		for _, q := range []string{
			`DELETE FROM payments WHERE order_id = ?`,
			`DELETE FROM order_status_logs WHERE order_id = ?`,
			`DELETE FROM order_items WHERE order_id = ?`,
			`DELETE FROM orders WHERE id = ?`,
		} {
			db.Exec(q, orderID)
		}
	})

	want := []int64{20000, 24000, 27500, 15000}
	if len(order.Items) != len(want) {
		t.Fatalf("jumlah item %d, want %d", len(order.Items), len(want))
	}
	for i, w := range want {
		if order.Items[i].Price != w {
			t.Errorf("item %d: harga %d, want %d", i, order.Items[i].Price, w)
		}
	}
	// 20000 + 2*24000 + 27500 + 15000 = 110500 ; subtotal tersimpan di database
	var subtotal int64
	if err := db.QueryRow(`SELECT subtotal FROM orders WHERE id = ?`, orderID).Scan(&subtotal); err != nil {
		t.Fatal(err)
	}
	if subtotal != 110500 {
		t.Errorf("subtotal %d, want 110500", subtotal)
	}
}

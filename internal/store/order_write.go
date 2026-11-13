package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type NewOrderItem struct {
	MenuItemID *int64 // bila ada, nama/kategori/harga diambil dari database
	Name       string
	Category   string
	Price      int64
	Qty        int64
	Size       *string
	Ice        *string // Normal | Less | No
	Sugar      *string // Normal | Less | No
}

type NewOrder struct {
	CustomerName  string
	OrderType     string
	TableNumber   *string
	PaymentMethod string
	CashReceived  *int64
	Note          string
	Items         []NewOrderItem
}

type OrderUpdate struct {
	Status       *string
	Paid         *bool
	CashReceived *int64
	CancelReason *string
}

// nextStatus: alur maju pesanan. Pembatalan ditangani terpisah.
var nextStatus = map[string]string{
	"menunggu": "proses",
	"proses":   "siap",
	"siap":     "selesai",
}

const qrisProvider = "manual"

// PPNPercent adalah tarif PPN (persen). Harus sama dengan PPN_PERCENT di frontend (shared/utils/orderTotals.js).
const PPNPercent = 11

// maxSubtotal menjaga subtotal + PPN tetap muat di kolom orders.total (INT UNSIGNED).
const maxSubtotal = 3_800_000_000

// CalcPPN menghitung PPN dengan pembulatan ke rupiah terdekat (setengah naik),
// sama dengan Math.round(subtotal * 0.11) di frontend.
func CalcPPN(subtotal int64) int64 {
	return (subtotal*PPNPercent + 50) / 100
}

// Create membuat pesanan dari kasir: status proses, source Kasir, langsung lunas
// (sama dengan perilaku halaman Buat Pesanan).
func (s *OrderStore) Create(ctx context.Context, in NewOrder, cashierID int64) (*Order, error) {
	return s.create(ctx, in, createOpts{
		Source: "Kasir", Status: "proses", Paid: true, CashierID: &cashierID, ChangedBy: cashierID,
	})
}

// CreateMobile membuat pesanan dari aplikasi customer: status menunggu, belum lunas,
// belum ada kasir. Pembayaran dikonfirmasi kasir lewat PUT /kasir/pesanan/:id.
func (s *OrderStore) CreateMobile(ctx context.Context, in NewOrder, customerID int64, phone *string) (*Order, error) {
	return s.create(ctx, in, createOpts{
		Source: "Mobile", Status: "menunggu", CustomerID: &customerID, Phone: phone, ChangedBy: customerID,
	})
}

type createOpts struct {
	Source     string
	Status     string
	Paid       bool
	CashierID  *int64
	CustomerID *int64
	Phone      *string
	ChangedBy  int64
}

// create menjalankan semua langkah pembuatan pesanan dalam satu transaksi.
func (s *OrderStore) create(ctx context.Context, in NewOrder, o createOpts) (*Order, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var subtotal int64
	items := make([]NewOrderItem, 0, len(in.Items))
	for _, it := range in.Items {
		resolved, err := resolveItem(ctx, tx, it)
		if err != nil {
			return nil, err
		}
		subtotal += resolved.Price * resolved.Qty
		items = append(items, resolved)
	}
	if subtotal <= 0 || subtotal > maxSubtotal {
		return nil, ValidationError("Total pesanan tidak valid")
	}
	// Total yang disimpan = subtotal + PPN (diskon belum didukung di POST /kasir/pesanan).
	total := subtotal + CalcPPN(subtotal)

	var cash any
	if o.Paid && in.PaymentMethod == "Tunai (Cash)" && in.CashReceived != nil {
		if *in.CashReceived < total {
			return nil, ValidationError("Uang diterima kurang dari total pesanan")
		}
		cash = *in.CashReceived
	}

	name := strings.TrimSpace(in.CustomerName)
	if name == "" {
		name = "Pelanggan"
	}

	res, err := tx.ExecContext(ctx,
		`INSERT INTO orders (customer_id, customer_name, phone, source, cashier_id, cashier_name, order_type, table_number,
		                     status, payment_method, paid, paid_at, cash_received, subtotal, discount, total, note)
		 VALUES (?, ?, ?, ?, ?, (SELECT name FROM users WHERE id = ?), ?, ?, ?, ?, ?, IF(?, NOW(), NULL), ?, ?, 0, ?, ?)`,
		o.CustomerID, name, o.Phone, o.Source, o.CashierID, o.CashierID, in.OrderType, in.TableNumber,
		o.Status, in.PaymentMethod, o.Paid, o.Paid, cash, subtotal, total, nullIfEmpty(in.Note))
	if err != nil {
		return nil, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	for _, it := range items {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, menu_item_id, name, category, price, qty, size, ice, sugar)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			orderID, it.MenuItemID, it.Name, it.Category, it.Price, it.Qty, it.Size, it.Ice, it.Sugar); err != nil {
			return nil, err
		}
	}
	if err := logStatus(ctx, tx, orderID, nil, o.Status, o.ChangedBy); err != nil {
		return nil, err
	}
	if in.PaymentMethod == "QRIS" {
		if err := recordQRIS(ctx, tx, orderID, total, o.Paid); err != nil {
			return nil, err
		}
	}

	order, err := getOrder(ctx, tx, orderID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

// resolveItem melengkapi dan memvalidasi satu item pesanan.
func resolveItem(ctx context.Context, tx *sql.Tx, it NewOrderItem) (NewOrderItem, error) {
	var customizable bool
	var priceLarge sql.NullInt64 // harga Large dari database (NULL = belum diisi)
	fromMenu := it.MenuItemID != nil

	if fromMenu {
		var status string
		err := tx.QueryRowContext(ctx,
			`SELECT m.name, c.name, m.price, m.price_large, m.status, c.is_customizable
			   FROM menu_items m JOIN menu_categories c ON c.id = m.category_id
			  WHERE m.id = ? AND m.deleted_at IS NULL`, *it.MenuItemID,
		).Scan(&it.Name, &it.Category, &it.Price, &priceLarge, &status, &customizable)
		if errors.Is(err, sql.ErrNoRows) {
			return it, ValidationError(fmt.Sprintf("Menu dengan id %d tidak ditemukan", *it.MenuItemID))
		}
		if err != nil {
			return it, err
		}
		if status != "tersedia" {
			return it, ValidationError(fmt.Sprintf("Menu %s sedang habis", it.Name))
		}
	} else {
		err := tx.QueryRowContext(ctx,
			`SELECT is_customizable FROM menu_categories WHERE name = ?`, it.Category).Scan(&customizable)
		if errors.Is(err, sql.ErrNoRows) {
			return it, ValidationError(fmt.Sprintf("Kategori %q tidak valid", it.Category))
		}
		if err != nil {
			return it, err
		}
	}

	if customizable {
		if it.Size == nil {
			regular := "Regular"
			it.Size = &regular
		}
		if it.Ice == nil {
			normal := "Normal"
			it.Ice = &normal
		}
		if it.Sugar == nil {
			normal := "Normal"
			it.Sugar = &normal
		}
	} else {
		it.Size, it.Ice, it.Sugar = nil, nil, nil
	}

	// Harga satuan mengikuti ukuran. Hanya untuk item dari menu (harga dari database);
	// item manual dari kasir (tanpa menuItemId) dipercaya apa adanya.
	if fromMenu {
		var large *int64
		if priceLarge.Valid {
			large = &priceLarge.Int64
		}
		it.Price = UnitPrice(it.Price, large, it.Size)
	}
	return it, nil
}

// Update mengubah status, menandai lunas, atau membatalkan pesanan.
// Aturan: alur menunggu -> proses -> siap -> selesai (satu langkah), dibatalkan dari tahap
// mana pun sebelum selesai, selesai hanya bila sudah lunas, pesanan final tidak bisa diubah.
func (s *OrderStore) Update(ctx context.Context, id, cashierID int64, in OrderUpdate) (*Order, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var status, method string
	var paid bool
	var total int64
	var owner sql.NullInt64
	err = tx.QueryRowContext(ctx,
		`SELECT status, paid, payment_method, total, cashier_id
		   FROM orders WHERE id = ? FOR UPDATE`, id).Scan(&status, &paid, &method, &total, &owner)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == "selesai" || status == "dibatalkan" {
		return nil, ValidationError("Pesanan yang sudah selesai atau dibatalkan tidak dapat diubah")
	}

	sets := make([]string, 0, 6)
	args := make([]any, 0, 6)
	payNow := false

	// Pembayaran
	if in.Paid != nil {
		switch {
		case !*in.Paid && paid:
			return nil, ValidationError("Pembayaran yang sudah lunas tidak dapat dibatalkan")
		case *in.Paid && !paid:
			payNow = true
			sets = append(sets, "paid = 1", "paid_at = NOW()")
			if method == "Tunai (Cash)" {
				if in.CashReceived == nil {
					return nil, ValidationError("Uang diterima wajib diisi untuk pembayaran tunai")
				}
				if *in.CashReceived < total {
					return nil, ValidationError("Uang diterima kurang dari total pesanan")
				}
				sets = append(sets, "cash_received = ?")
				args = append(args, *in.CashReceived)
			}
		}
	}
	if in.CashReceived != nil && !payNow {
		return nil, ValidationError("Uang diterima hanya dikirim bersama penandaan lunas")
	}

	// Status
	newStatus := status
	if in.Status != nil && *in.Status != status {
		target := *in.Status
		switch {
		case target == "dibatalkan":
			reason := ""
			if in.CancelReason != nil {
				reason = strings.TrimSpace(*in.CancelReason)
			}
			if reason == "" {
				return nil, ValidationError("Alasan pembatalan wajib diisi")
			}
			sets = append(sets, "cancel_reason = ?", "cancelled_at = NOW()")
			args = append(args, reason)
		case nextStatus[status] == target:
			if target == "selesai" && !paid && !payNow {
				return nil, ValidationError("Pesanan belum lunas, tandai lunas sebelum menyelesaikan")
			}
		default:
			return nil, ValidationError(fmt.Sprintf("Status tidak dapat diubah dari %s ke %s", status, target))
		}
		newStatus = target
		sets = append(sets, "status = ?")
		args = append(args, target)
	}

	if len(sets) == 0 {
		o, err := getOrder(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		return o, nil // tidak ada perubahan
	}

	// Pesanan dari mobile dipegang kasir pertama yang memprosesnya.
	if !owner.Valid {
		// cashier_name disimpan sebagai salinan nama agar riwayat tetap terbaca walau akun kasir dihapus permanen.
		sets = append(sets, "cashier_id = ?", "cashier_name = (SELECT name FROM users WHERE id = ?)")
		args = append(args, cashierID, cashierID)
	}

	args = append(args, id)
	if _, err := tx.ExecContext(ctx, `UPDATE orders SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...); err != nil {
		return nil, err
	}

	if newStatus != status {
		if err := logStatus(ctx, tx, id, &status, newStatus, cashierID); err != nil {
			return nil, err
		}
	}
	if payNow && method == "QRIS" {
		if err := recordQRIS(ctx, tx, id, total, true); err != nil {
			return nil, err
		}
	}
	if newStatus == "dibatalkan" {
		if _, err := tx.ExecContext(ctx,
			`UPDATE payments SET status = 'refunded', refunded_at = NOW()
			  WHERE order_id = ? AND status = 'paid'`, id); err != nil {
			return nil, err
		}
	}

	o, err := getOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return o, nil
}

func logStatus(ctx context.Context, tx *sql.Tx, orderID int64, from *string, to string, by int64) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO order_status_logs (order_id, from_status, to_status, changed_by) VALUES (?, ?, ?, ?)`,
		orderID, from, to, by)
	return err
}

// recordQRIS mencatat pembayaran QRIS (dikonfirmasi kasir secara manual, belum ada gateway).
// paid=false membuat baris pending; konfirmasi kasir kemudian mengubahnya jadi paid.
func recordQRIS(ctx context.Context, tx *sql.Tx, orderID, amount int64, paid bool) error {
	status := "pending"
	if paid {
		status = "paid"
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO payments (order_id, provider, provider_ref, amount, status, paid_at)
		 VALUES (?, ?, ?, ?, ?, IF(?, NOW(), NULL))
		 ON DUPLICATE KEY UPDATE status = VALUES(status), paid_at = VALUES(paid_at), refunded_at = NULL`,
		orderID, qrisProvider, fmt.Sprintf("ORDER-%d", orderID), amount, status, paid)
	return err
}

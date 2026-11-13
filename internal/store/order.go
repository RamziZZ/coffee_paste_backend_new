package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"coffeepaste/internal/format"
)

type OrderStore struct {
	DB *sql.DB
}

type OrderItem struct {
	MenuItemID *int64  `json:"menuItemId"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Price      int64   `json:"price"`
	Qty        int64   `json:"qty"`
	Size       *string `json:"size"`
	Ice        *string `json:"ice"`
	Sugar      *string `json:"sugar"`
	Image      *string `json:"image"`
}

// Order adalah bentuk pesanan pada respons API. Subtotal dan total sengaja tidak dikirim
// (frontend menghitungnya dari items dan discount), tetapi tetap disimpan di database.
type Order struct {
	ID            string      `json:"id"` // kode #CP00001
	Date          string      `json:"date"`
	CustomerName  string      `json:"customerName"`
	Phone         *string     `json:"phone"`
	Source        string      `json:"source"`
	CashierID     *string     `json:"cashierId"`
	CashierName   *string     `json:"cashierName"`
	OrderType     string      `json:"orderType"`
	TableNumber   *string     `json:"tableNumber"`
	Status        string      `json:"status"`
	PaymentMethod string      `json:"paymentMethod"`
	Paid          bool        `json:"paid"`
	CashReceived  *int64      `json:"cashReceived"`
	Items         []OrderItem `json:"items"`
	Discount      int64       `json:"discount"`
	CancelReason  *string     `json:"cancelReason"`
	Note          string      `json:"note"`
}

const orderSelect = `SELECT o.id, o.code, o.created_at, o.customer_name, o.phone, o.source,
	       o.cashier_id, o.cashier_name, o.order_type, o.table_number, o.status,
	       o.payment_method, o.paid, o.cash_received, o.discount, o.cancel_reason, COALESCE(o.note, '')
	  FROM v_orders o`

// scanOrder membaca satu baris orderSelect. Mengembalikan id numerik terpisah untuk memuat item.
func scanOrder(r rowScanner) (Order, int64, error) {
	var o Order
	var id int64
	var created time.Time
	var phone, cashierName, table, cancel sql.NullString
	var cashierID, cash sql.NullInt64
	var paid bool
	err := r.Scan(&id, &o.ID, &created, &o.CustomerName, &phone, &o.Source,
		&cashierID, &cashierName, &o.OrderType, &table, &o.Status,
		&o.PaymentMethod, &paid, &cash, &o.Discount, &cancel, &o.Note)
	if err != nil {
		return o, 0, err
	}
	o.Date = created.Format(format.ISO)
	o.Phone = nullStr(phone)
	o.CashierName = nullStr(cashierName)
	o.TableNumber = nullStr(table)
	o.CancelReason = nullStr(cancel)
	o.Paid = paid
	if cashierID.Valid {
		v := strconv.FormatInt(cashierID.Int64, 10)
		o.CashierID = &v
	}
	if cash.Valid {
		v := cash.Int64
		o.CashReceived = &v
	}
	o.Items = make([]OrderItem, 0)
	return o, id, nil
}

// OrderFilter membatasi daftar pesanan. Field kosong berarti tidak difilter.
type OrderFilter struct {
	From       string // YYYY-MM-DD, inklusif
	To         string // YYYY-MM-DD, inklusif
	CustomerID int64  // 0 = semua customer
}

// where menyusun klausa WHERE (alias tabel orders: o) beserta argumennya.
func (f OrderFilter) where() (string, []any) {
	var conds []string
	var args []any
	if f.From != "" {
		conds = append(conds, "o.created_at >= ?")
		args = append(args, f.From)
	}
	if f.To != "" {
		conds = append(conds, "o.created_at < DATE_ADD(?, INTERVAL 1 DAY)")
		args = append(args, f.To)
	}
	if f.CustomerID > 0 {
		conds = append(conds, "o.customer_id = ?")
		args = append(args, f.CustomerID)
	}
	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// List mengembalikan pesanan (terbaru dulu) lengkap dengan item.
func (s *OrderStore) List(ctx context.Context, f OrderFilter) ([]Order, error) {
	where, args := f.where()
	rows, err := s.DB.QueryContext(ctx, orderSelect+where+` ORDER BY o.created_at DESC, o.id DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	index := make(map[int64]int)
	for rows.Next() {
		o, id, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		index[id] = len(orders)
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	if len(orders) == 0 {
		return orders, nil
	}
	if err := s.attachItems(ctx, orders, index, f); err != nil {
		return nil, err
	}
	return orders, nil
}

// Get mengambil satu pesanan berdasarkan id numerik.
func (s *OrderStore) Get(ctx context.Context, id int64) (*Order, error) {
	return getOrder(ctx, s.DB, id)
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getOrder(ctx context.Context, q queryer, id int64) (*Order, error) {
	o, _, err := scanOrder(q.QueryRowContext(ctx, orderSelect+` WHERE o.id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	rows, err := q.QueryContext(ctx, itemSelect+` WHERE oi.order_id = ? ORDER BY oi.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		_, it, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}

const itemSelect = `SELECT oi.order_id, oi.menu_item_id, oi.name, oi.category, oi.price, oi.qty,
	       oi.size, oi.ice, oi.sugar, m.image_url
	  FROM order_items oi
	  LEFT JOIN menu_items m ON m.id = oi.menu_item_id`

func scanItem(r rowScanner) (int64, OrderItem, error) {
	var orderID int64
	var it OrderItem
	var menuID sql.NullInt64
	var size, ice, sugar, img sql.NullString
	if err := r.Scan(&orderID, &menuID, &it.Name, &it.Category, &it.Price, &it.Qty,
		&size, &ice, &sugar, &img); err != nil {
		return 0, it, err
	}
	if menuID.Valid {
		v := menuID.Int64
		it.MenuItemID = &v
	}
	it.Size = nullStr(size)
	it.Ice = nullStr(ice)
	it.Sugar = nullStr(sugar)
	it.Image = nullStr(img)
	if it.Image != nil && *it.Image == "" {
		it.Image = nil
	}
	return orderID, it, nil
}

// attachItems memuat item untuk semua pesanan dalam satu query tambahan.
func (s *OrderStore) attachItems(ctx context.Context, orders []Order, index map[int64]int, f OrderFilter) error {
	where, args := f.where()
	q := itemSelect
	if where != "" {
		q += ` JOIN orders o ON o.id = oi.order_id` + where
	}
	rows, err := s.DB.QueryContext(ctx, q+` ORDER BY oi.order_id, oi.id`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		orderID, it, err := scanItem(rows)
		if err != nil {
			return err
		}
		if i, ok := index[orderID]; ok {
			orders[i].Items = append(orders[i].Items, it)
		}
	}
	return rows.Err()
}

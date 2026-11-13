package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/realtime"
	"coffeepaste/internal/store"
)

type KasirOrderHandler struct {
	orders *store.OrderStore

	hub      *realtime.Hub // siaran realtime (SSE); nil = nonaktif
	notifier *orderNotifier
}

func NewKasirOrderHandler(orders *store.OrderStore) *KasirOrderHandler {
	return &KasirOrderHandler{orders: orders}
}

// orderPathID menerima "#CP00001", "CP00001", atau id numerik "1".
// Catatan: karakter "#" harus dikirim sebagai %23 (encodeURIComponent), kalau tidak
// browser memperlakukannya sebagai fragment dan tidak ikut terkirim.
func orderPathID(r *http.Request) (int64, bool) {
	return parseOrderCode(r.PathValue("id"))
}

// parseOrderCode mengubah "#CP00001", "CP00001", atau "1" menjadi id numerik.
func parseOrderCode(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "#")
	if len(raw) >= 2 && strings.EqualFold(raw[:2], "CP") {
		raw = raw[2:]
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	return id, err == nil && id > 0
}

// List: GET /api/kasir/pesanan
func (h *KasirOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	orders, err := h.orders.List(r.Context(), store.OrderFilter{})
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, orders)
}

type orderItemRequest struct {
	MenuItemID *int64  `json:"menuItemId"`
	Name       string  `json:"name"`
	Category   string  `json:"category"`
	Price      int64   `json:"price"`
	Qty        int64   `json:"qty"`
	Size       *string `json:"size"`
	Ice        *string `json:"ice"`
	Sugar      *string `json:"sugar"`
	NoIce      *bool   `json:"noIce"`     // lama, hanya dibaca bila "ice" kosong
	LessSugar  *bool   `json:"lessSugar"` // lama, hanya dibaca bila "sugar" kosong
}

func validLevel(v string) bool { return v == "Normal" || v == "Less" || v == "No" }

// legacyLevel memetakan flag boolean lama ke tingkat baru.
func legacyLevel(flag *bool, on string) *string {
	if flag == nil {
		return nil
	}
	v := "Normal"
	if *flag {
		v = on
	}
	return &v
}

type createOrderRequest struct {
	CustomerName  string             `json:"customerName"`
	OrderType     string             `json:"orderType"`
	TableNumber   *string            `json:"tableNumber"`
	PaymentMethod string             `json:"paymentMethod"`
	CashReceived  *int64             `json:"cashReceived"`
	Note          string             `json:"note"`
	Items         []orderItemRequest `json:"items"`
}

func (q createOrderRequest) toInput() (store.NewOrder, string) {
	name := strings.TrimSpace(q.CustomerName)
	note := strings.TrimSpace(q.Note)

	switch {
	case utf8.RuneCountInString(name) > 100:
		return store.NewOrder{}, "Nama pelanggan maksimal 100 karakter"
	case q.OrderType != "Dine In" && q.OrderType != "Take Away":
		return store.NewOrder{}, "Tipe pesanan harus Dine In atau Take Away"
	case q.PaymentMethod != "QRIS" && q.PaymentMethod != "Tunai (Cash)":
		return store.NewOrder{}, "Metode pembayaran harus QRIS atau Tunai (Cash)"
	case q.CashReceived != nil && (*q.CashReceived < 0 || *q.CashReceived > maxRupiah):
		return store.NewOrder{}, "Uang diterima tidak valid"
	case len(q.Items) == 0:
		return store.NewOrder{}, "Pesanan minimal berisi satu item"
	case len(q.Items) > 100:
		return store.NewOrder{}, "Pesanan maksimal berisi 100 item"
	case utf8.RuneCountInString(note) > 1000:
		return store.NewOrder{}, "Catatan maksimal 1000 karakter"
	}

	var table *string
	if q.TableNumber != nil {
		if v := strings.TrimSpace(*q.TableNumber); v != "" {
			if q.OrderType == "Take Away" {
				return store.NewOrder{}, "Nomor meja harus dikosongkan untuk Take Away"
			}
			if utf8.RuneCountInString(v) > 20 {
				return store.NewOrder{}, "Nomor meja maksimal 20 karakter"
			}
			table = &v
		}
	}

	items := make([]store.NewOrderItem, 0, len(q.Items))
	for i, it := range q.Items {
		n := i + 1
		itName := strings.TrimSpace(it.Name)
		itCat := strings.TrimSpace(it.Category)

		if it.Qty < 1 || it.Qty > 999 {
			return store.NewOrder{}, "Jumlah item ke-" + strconv.Itoa(n) + " harus antara 1 dan 999"
		}
		if it.Size != nil && *it.Size != "Regular" && *it.Size != "Large" {
			return store.NewOrder{}, "Ukuran item ke-" + strconv.Itoa(n) + " harus Regular atau Large"
		}
		ice, sugar := it.Ice, it.Sugar
		if ice == nil {
			ice = legacyLevel(it.NoIce, "No")
		}
		if sugar == nil {
			sugar = legacyLevel(it.LessSugar, "Less")
		}
		if ice != nil && !validLevel(*ice) {
			return store.NewOrder{}, "Es item ke-" + strconv.Itoa(n) + " harus Normal, Less, atau No"
		}
		if sugar != nil && !validLevel(*sugar) {
			return store.NewOrder{}, "Gula item ke-" + strconv.Itoa(n) + " harus Normal, Less, atau No"
		}
		if it.MenuItemID != nil {
			if *it.MenuItemID <= 0 {
				return store.NewOrder{}, "menuItemId item ke-" + strconv.Itoa(n) + " tidak valid"
			}
		} else {
			// Tanpa menuItemId, data item dipercaya apa adanya dari kasir.
			switch {
			case itName == "" || utf8.RuneCountInString(itName) > 100:
				return store.NewOrder{}, "Nama item ke-" + strconv.Itoa(n) + " wajib diisi (maksimal 100 karakter)"
			case itCat == "" || utf8.RuneCountInString(itCat) > 50:
				return store.NewOrder{}, "Kategori item ke-" + strconv.Itoa(n) + " wajib diisi"
			case it.Price <= 0 || it.Price > maxRupiah:
				return store.NewOrder{}, "Harga item ke-" + strconv.Itoa(n) + " harus lebih dari 0"
			}
		}
		items = append(items, store.NewOrderItem{
			MenuItemID: it.MenuItemID, Name: itName, Category: itCat, Price: it.Price,
			Qty: it.Qty, Size: it.Size, Ice: ice, Sugar: sugar,
		})
	}

	return store.NewOrder{
		CustomerName:  name,
		OrderType:     q.OrderType,
		TableNumber:   table,
		PaymentMethod: q.PaymentMethod,
		CashReceived:  q.CashReceived,
		Note:          note,
		Items:         items,
	}, ""
}

// Create: POST /api/kasir/pesanan
func (h *KasirOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req createOrderRequest
	if !decodeOrFail(w, r, &req) {
		return
	}
	in, msg := req.toInput()
	if msg != "" {
		unprocessable(w, msg)
		return
	}

	order, err := h.orders.Create(r.Context(), in, user.ID)
	if h.failed(w, err) {
		return
	}
	h.notifier.changedByCode(r.Context(), order.ID)
	httpx.JSON(w, http.StatusCreated, order)
}

// Update: PUT /api/kasir/pesanan/{id}  (kirim hanya field yang berubah)
func (h *KasirOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	id, ok := orderPathID(r)
	if !ok {
		respondNotFound(w, "Pesanan tidak ditemukan")
		return
	}
	var req struct {
		Status       *string `json:"status"`
		Paid         *bool   `json:"paid"`
		CashReceived *int64  `json:"cashReceived"`
		CancelReason *string `json:"cancelReason"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}

	if req.Status != nil {
		switch *req.Status {
		case "menunggu", "proses", "siap", "selesai", "dibatalkan":
		default:
			unprocessable(w, "Status tidak valid")
			return
		}
	}
	if req.CashReceived != nil && (*req.CashReceived < 0 || *req.CashReceived > maxRupiah) {
		unprocessable(w, "Uang diterima tidak valid")
		return
	}
	if req.CancelReason != nil && utf8.RuneCountInString(*req.CancelReason) > 1000 {
		unprocessable(w, "Alasan pembatalan maksimal 1000 karakter")
		return
	}

	order, err := h.orders.Update(r.Context(), id, user.ID, store.OrderUpdate{
		Status: req.Status, Paid: req.Paid, CashReceived: req.CashReceived, CancelReason: req.CancelReason,
	})
	if h.failed(w, err) {
		return
	}
	h.notifier.changed(r.Context(), id)
	httpx.JSON(w, http.StatusOK, order)
}

// failed membalas error store yang umum dan mengembalikan true bila err != nil.
func (h *KasirOrderHandler) failed(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	var ve store.ValidationError
	switch {
	case errors.As(err, &ve):
		unprocessable(w, ve.Error())
	case errors.Is(err, store.ErrNotFound):
		respondNotFound(w, "Pesanan tidak ditemukan")
	default:
		httpx.ServerError(w, err)
	}
	return true
}

package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/midtrans"
	"coffeepaste/internal/qris"
	"coffeepaste/internal/realtime"
	"coffeepaste/internal/store"
)

// MobileHandler melayani endpoint aplikasi mobile (role customer).
type MobileHandler struct {
	reviews *store.ReviewStore
	menu    *store.MenuStore
	orders  *store.OrderStore
	users   *store.UserStore

	hub      *realtime.Hub // siaran realtime (SSE); nil = nonaktif
	notifier *orderNotifier

	qrisStatic string // kode QRIS statis merchant; kosong = QR tidak dibuat

	gateway    QRISCharger // payment gateway (Midtrans); nil = pakai QRIS statis
	chargeMu   sync.Mutex  // mencegah dua tagihan ganda untuk pesanan yang sama
	chargeFail map[int64]time.Time
	chargeMsg  map[int64]string // alasan singkat gagalnya tagihan terakhir (untuk ditampilkan di aplikasi)
}

// QRISCharger membuat tagihan QRIS di payment gateway.
type QRISCharger interface {
	ChargeQRIS(ctx context.Context, orderID string, amount int64, expiryMinutes int) (*midtrans.QRISCharge, error)
}

const (
	qrisExpiryMinutes = 30
	chargeRetryAfter  = 30 * time.Second // jeda coba ulang bila gateway gagal
)

// WithGateway mengaktifkan QRIS dinamis lewat payment gateway. Pesanan QRIS yang belum
// dibayar mendapat tagihan di gateway; bila gateway gagal, jatuh kembali ke QRIS statis.
func (h *MobileHandler) WithGateway(g QRISCharger) *MobileHandler {
	h.gateway = g
	return h
}

// WithQRIS mengaktifkan pembuatan QRIS dinamis pada pesanan QRIS yang belum dibayar.
func (h *MobileHandler) WithQRIS(static string) *MobileHandler {
	h.qrisStatic = static
	return h
}

func NewMobileHandler(reviews *store.ReviewStore, menu *store.MenuStore, orders *store.OrderStore, users *store.UserStore) *MobileHandler {
	return &MobileHandler{reviews: reviews, menu: menu, orders: orders, users: users, chargeFail: map[int64]time.Time{}, chargeMsg: map[int64]string{}}
}

func validationMsg(err error) (string, bool) {
	var ve store.ValidationError
	if errors.As(err, &ve) {
		return ve.Error(), true
	}
	return "", false
}

// CreateReview: POST /api/reviews  { rating, comment, orderId? }
// Aturan: orderId opsional; bila diisi harus pesanan milik customer yang berstatus selesai,
// dan satu pesanan hanya satu review. Review baru langsung tampil (is_published = 1).
func (h *MobileHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
		OrderID *int64 `json:"orderId"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	comment := strings.TrimSpace(req.Comment)
	switch {
	case req.Rating < 1 || req.Rating > 5:
		unprocessable(w, "Rating harus antara 1 sampai 5")
		return
	case comment == "":
		unprocessable(w, "Komentar wajib diisi")
		return
	case utf8.RuneCountInString(comment) > 2000:
		unprocessable(w, "Komentar maksimal 2000 karakter")
		return
	case req.OrderID != nil && *req.OrderID <= 0:
		unprocessable(w, "orderId tidak valid")
		return
	}

	id, err := h.reviews.Create(r.Context(), user.ID, req.OrderID, req.Rating, comment)
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"id": id, "rating": req.Rating, "comment": comment, "orderId": req.OrderID})
}

// Favorites: GET /api/favorites  -> [{ id, name, category, price, status, image, description }]
func (h *MobileHandler) Favorites(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	items, err := h.menu.Favorites(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

// AddFavorite: POST /api/favorites  { menuItemId }  (idempoten)
func (h *MobileHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	var req struct {
		MenuItemID int64 `json:"menuItemId"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if req.MenuItemID <= 0 {
		unprocessable(w, "menuItemId wajib diisi")
		return
	}
	item, err := h.menu.AddFavorite(r.Context(), user.ID, req.MenuItemID)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, item)
}

// RemoveFavorite: DELETE /api/favorites/{menuItemId}  (idempoten)
func (h *MobileHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(r.PathValue("menuItemId"), 10, 64)
	if err != nil || id <= 0 {
		respondNotFound(w, "Menu tidak ditemukan")
		return
	}
	if err := h.menu.RemoveFavorite(r.Context(), user.ID, id); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// CreateOrder: POST /api/orders
// Body sama dengan POST /kasir/pesanan, tetapi setiap item wajib memakai menuItemId (harga
// selalu dari database) dan cashReceived diabaikan. Pesanan masuk berstatus menunggu, belum
// lunas, source Mobile; nama dan telepon diambil dari akun customer.
func (h *MobileHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
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
	for i, it := range in.Items {
		if it.MenuItemID == nil {
			unprocessable(w, "menuItemId item ke-"+strconv.Itoa(i+1)+" wajib diisi")
			return
		}
	}
	in.CashReceived = nil
	in.CustomerName = user.Name

	phone, err := h.users.Phone(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	order, err := h.orders.CreateMobile(r.Context(), in, user.ID, phone)
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	h.notifier.changedByCode(r.Context(), order.ID)
	httpx.JSON(w, http.StatusCreated, h.mobileOrderCtx(r.Context(), *order))
}

// SaveDevice: POST /api/devices  { token, platform: "android" | "ios" }
func (h *MobileHandler) SaveDevice(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	var req struct {
		Token    string `json:"token"`
		Platform string `json:"platform"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	token := strings.TrimSpace(req.Token)
	switch {
	case token == "" || len(token) > 255:
		unprocessable(w, "Token perangkat wajib diisi (maksimal 255 karakter)")
		return
	case req.Platform != "android" && req.Platform != "ios":
		unprocessable(w, "Platform harus android atau ios")
		return
	}
	if err := h.users.SaveDeviceToken(r.Context(), user.ID, token, req.Platform); err != nil {
		httpx.ServerError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// mobileOrder = pesanan + angka ringkasan. Respons kasir/admin tidak membawa total (frontend web
// menghitungnya sendiri), sedangkan aplikasi mobile menampilkan total apa adanya dari server
// sehingga PPN selalu sama dengan yang tersimpan di database.
type mobileOrder struct {
	store.Order
	Subtotal int64 `json:"subtotal"`
	Tax      int64 `json:"tax"`
	Total    int64 `json:"total"`
	// ReviewedItemIDs: id menu pada pesanan ini yang sudah diulas customer (untuk menyembunyikan
	// tombol "Beri Ulasan" pada produk yang sudah diulas).
	ReviewedItemIDs []int64 `json:"reviewedItemIds"`
	// QRISPayload: isi QR (QRIS dinamis berisi total) untuk pesanan QRIS yang belum lunas dan
	// belum dibatalkan. Kosong bila bukan QRIS, sudah lunas, atau QRIS merchant belum dikonfigurasi.
	QRISPayload string `json:"qrisPayload,omitempty"`
	// QRISError: alasan QR belum bisa ditampilkan (gateway menolak, QRIS belum dikonfigurasi, dst).
	// Hanya terisi untuk pesanan QRIS yang belum lunas dan QR-nya kosong.
	QRISError string `json:"qrisError,omitempty"`
}

func toMobileOrder(o store.Order) mobileOrder {
	var subtotal int64
	for _, it := range o.Items {
		subtotal += it.Price * it.Qty
	}
	tax := store.CalcPPN(subtotal)
	return mobileOrder{Order: o, Subtotal: subtotal, Tax: tax, Total: subtotal + tax - o.Discount, ReviewedItemIDs: []int64{}}
}

// mobileOrder membangun respons pesanan untuk mobile, termasuk QR pembayaran bila perlu.
func (h *MobileHandler) mobileOrder(o store.Order) mobileOrder {
	m := toMobileOrder(o)
	if h.qrisStatic != "" && o.PaymentMethod == "QRIS" && !o.Paid && o.Status != "dibatalkan" && m.Total > 0 {
		if code, err := qris.Dynamic(h.qrisStatic, m.Total); err == nil {
			m.QRISPayload = code
		}
	}
	return m
}

// mobileOrderCtx = mobileOrder + tagihan QRIS dari payment gateway (bila aktif). Bila gateway
// tidak aktif atau gagal, QRISPayload tetap berisi QRIS statis dinamis (jika dikonfigurasi).
func (h *MobileHandler) mobileOrderCtx(ctx context.Context, o store.Order) mobileOrder {
	m := h.mobileOrder(o)
	if o.PaymentMethod != "QRIS" || o.Paid || o.Status == "dibatalkan" || m.Total <= 0 {
		return m
	}
	if h.gateway != nil && h.orders != nil {
		if id, ok := parseOrderCode(o.ID); ok {
			if qr := h.gatewayQR(ctx, id, m.Total); qr != "" {
				m.QRISPayload = qr
			} else if m.QRISPayload == "" {
				m.QRISError = h.lastChargeMsg(id)
			}
		}
	}
	if m.QRISPayload == "" && m.QRISError == "" {
		m.QRISError = "QRIS belum dikonfigurasi di server (isi MIDTRANS_SERVER_KEY atau QRIS_STATIC)."
	}
	return m
}

// lastChargeMsg mengembalikan alasan gagalnya tagihan terakhir untuk pesanan itu.
func (h *MobileHandler) lastChargeMsg(orderID int64) string {
	h.chargeMu.Lock()
	defer h.chargeMu.Unlock()
	if msg := h.chargeMsg[orderID]; msg != "" {
		return msg
	}
	return "Kode QR belum bisa dibuat. Coba muat ulang sebentar lagi."
}

// chargeErrorMessage menerjemahkan error gateway menjadi pesan singkat yang aman ditampilkan
// ke customer/kasir (detail teknis tetap hanya di log server).
func chargeErrorMessage(err error) string {
	e := strings.ToLower(err.Error())
	switch {
	case strings.Contains(e, "401") || strings.Contains(e, "unauthorized"):
		return "Kunci Midtrans ditolak. Periksa MIDTRANS_SERVER_KEY dan MIDTRANS_IS_PRODUCTION di server."
	case strings.Contains(e, "402") || strings.Contains(e, "not activated") || strings.Contains(e, "payment channel"):
		return "Channel QRIS belum aktif di akun Midtrans."
	case strings.Contains(e, "timeout") || strings.Contains(e, "deadline") || strings.Contains(e, "no such host") || strings.Contains(e, "connection"):
		return "Server tidak bisa menghubungi Midtrans. Periksa koneksi internet server."
	default:
		detail := strings.TrimSpace(strings.TrimPrefix(err.Error(), "midtrans:"))
		if r := []rune(detail); len(r) > 160 {
			detail = string(r[:160]) + "..."
		}
		return "Midtrans menolak pembuatan QR: " + detail
	}
}

// gatewayQR mengembalikan QR tagihan yang masih berlaku, atau membuat tagihan baru di gateway.
// Dikunci agar polling bersamaan dari aplikasi tidak membuat tagihan ganda.
func (h *MobileHandler) gatewayQR(ctx context.Context, orderID, total int64) string {
	h.chargeMu.Lock()
	defer h.chargeMu.Unlock()

	qr, err := h.orders.ActiveMidtransQR(ctx, orderID)
	if err != nil {
		log.Printf("QRIS pesanan %d: gagal membaca tagihan: %v", orderID, err)
		return ""
	}
	if qr != "" {
		return qr
	}
	if t, ok := h.chargeFail[orderID]; ok && time.Since(t) < chargeRetryAfter {
		return ""
	}

	ref := fmt.Sprintf("CP%05d-%d", orderID, time.Now().Unix())
	ch, err := h.gateway.ChargeQRIS(ctx, ref, total, qrisExpiryMinutes)
	if err != nil {
		h.chargeFail[orderID] = time.Now()
		h.chargeMsg[orderID] = chargeErrorMessage(err)
		log.Printf("QRIS pesanan %d: gagal membuat tagihan: %v", orderID, err)
		return ""
	}
	if err := h.orders.SaveMidtransCharge(ctx, orderID, ref, ch.QRString, total, qrisExpiryMinutes); err != nil {
		h.chargeFail[orderID] = time.Now()
		h.chargeMsg[orderID] = "Tagihan dibuat tetapi gagal disimpan di database. Lihat log server."
		log.Printf("QRIS pesanan %d: gagal menyimpan tagihan: %v", orderID, err)
		return ""
	}
	delete(h.chargeFail, orderID)
	delete(h.chargeMsg, orderID)
	log.Printf("Tagihan QRIS dibuat: ref=%s total=%d qr_url=%s", ref, total, ch.QRURL)
	log.Printf("  tes lokal tanpa Midtrans: go run ./cmd/midtrans-sim -ref %s -amount %d", ref, total)
	return ch.QRString
}

// Orders: GET /api/orders -> pesanan milik customer yang login, terbaru dulu.
func (h *MobileHandler) Orders(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	orders, err := h.orders.List(r.Context(), store.OrderFilter{CustomerID: user.ID})
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	reviewed, err := h.reviews.ReviewedItems(r.Context(), user.ID)
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	out := make([]mobileOrder, 0, len(orders))
	for _, o := range orders {
		m := h.mobileOrderCtx(r.Context(), o)
		if id, ok := parseOrderCode(o.ID); ok && reviewed[id] != nil {
			m.ReviewedItemIDs = reviewed[id]
		}
		out = append(out, m)
	}
	httpx.JSON(w, http.StatusOK, out)
}

// CancelOrder: POST /api/orders/{id}/cancel
// Hanya pesanan milik sendiri yang masih berstatus menunggu dan belum lunas (lainnya 422).
// {id} boleh "#CP00001" (kirim %23CP00001), "CP00001", atau "1".
func (h *MobileHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	id, ok := orderPathID(r)
	if !ok {
		respondNotFound(w, "Pesanan tidak ditemukan")
		return
	}
	order, err := h.orders.CancelByCustomer(r.Context(), id, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		respondNotFound(w, "Pesanan tidak ditemukan")
		return
	}
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}
	h.notifier.changed(r.Context(), id)
	httpx.JSON(w, http.StatusOK, h.mobileOrderCtx(r.Context(), *order))
}

type productReviewRequest struct {
	MenuItemID int64  `json:"menuItemId"`
	Rating     int    `json:"rating"`
	Comment    string `json:"comment"`
}

// ReviewOrder: POST /api/orders/{id}/reviews
// Body: { "reviews": [ { "menuItemId": 3, "rating": 5, "comment": "Enak" } ] }
// Ulasan per produk dari pesanan yang sudah selesai. Rating 1-5 wajib, komentar boleh kosong.
// Satu produk hanya bisa diulas sekali per pesanan. Semua atau tidak sama sekali (422).
// Ulasan langsung tampil (is_published = 1); admin bisa menyembunyikannya.
func (h *MobileHandler) ReviewOrder(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFromContext(r.Context())
	orderID, ok := orderPathID(r)
	if !ok {
		respondNotFound(w, "Pesanan tidak ditemukan")
		return
	}

	var req struct {
		Reviews []productReviewRequest `json:"reviews"`
	}
	if !decodeOrFail(w, r, &req) {
		return
	}
	if len(req.Reviews) == 0 || len(req.Reviews) > 50 {
		unprocessable(w, "Kirim 1 sampai 50 ulasan")
		return
	}

	in := make([]store.ProductReviewInput, 0, len(req.Reviews))
	for i, rv := range req.Reviews {
		comment := strings.TrimSpace(rv.Comment)
		switch {
		case rv.MenuItemID <= 0:
			unprocessable(w, "menuItemId ulasan ke-"+strconv.Itoa(i+1)+" tidak valid")
			return
		case rv.Rating < 1 || rv.Rating > 5:
			unprocessable(w, "Rating harus antara 1 sampai 5")
			return
		case utf8.RuneCountInString(comment) > 2000:
			unprocessable(w, "Komentar maksimal 2000 karakter")
			return
		}
		in = append(in, store.ProductReviewInput{MenuItemID: rv.MenuItemID, Rating: rv.Rating, Comment: comment})
	}

	ids, err := h.reviews.CreateForOrder(r.Context(), user.ID, orderID, in)
	if msg, ok := validationMsg(err); ok {
		unprocessable(w, msg)
		return
	}
	if err != nil {
		httpx.ServerError(w, err)
		return
	}

	out := make([]map[string]any, 0, len(ids))
	for i, id := range ids {
		out = append(out, map[string]any{
			"id": id, "menuItemId": in[i].MenuItemID, "rating": in[i].Rating, "comment": in[i].Comment,
		})
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"reviews": out})
}

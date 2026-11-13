package handler

import (
	"fmt"
	"net/http"
	"time"

	"coffeepaste/internal/httpx"
	"coffeepaste/internal/middleware"
	"coffeepaste/internal/realtime"
)

// sseHeartbeat = jeda kiriman komentar ": ping" agar proxy/load balancer tidak menutup koneksi yang sepi.
const sseHeartbeat = 20 * time.Second

// WithHub mengaktifkan siaran realtime (SSE) untuk aplikasi mobile.
func (h *MobileHandler) WithHub(hub *realtime.Hub) *MobileHandler {
	h.hub = hub
	h.notifier = newOrderNotifier(h.orders, hub)
	return h
}

// WithHub mengaktifkan siaran realtime (SSE) untuk web kasir.
func (h *KasirOrderHandler) WithHub(hub *realtime.Hub) *KasirOrderHandler {
	h.hub = hub
	h.notifier = newOrderNotifier(h.orders, hub)
	return h
}

// WithHub membuat pelunasan lewat callback QRIS ikut disiarkan ke kasir dan customer.
func (h *PaymentCallbackHandler) WithHub(hub *realtime.Hub) *PaymentCallbackHandler {
	h.notifier = newOrderNotifier(h.orders, hub)
	return h
}

// WithHub membuat pelunasan lewat notifikasi Midtrans ikut disiarkan ke kasir dan customer.
func (h *MidtransHandler) WithHub(hub *realtime.Hub) *MidtransHandler {
	h.notifier = newOrderNotifier(h.orders, hub)
	return h
}

// WithPush mengaktifkan push notification (FCM). Panggil setelah WithHub.
func (h *MobileHandler) WithPush(p *OrderPusher) *MobileHandler {
	h.notifier.setPusher(p)
	return h
}

// WithPush mengaktifkan push notification (FCM). Panggil setelah WithHub.
func (h *KasirOrderHandler) WithPush(p *OrderPusher) *KasirOrderHandler {
	h.notifier.setPusher(p)
	return h
}

// WithPush mengaktifkan push notification (FCM). Panggil setelah WithHub.
func (h *PaymentCallbackHandler) WithPush(p *OrderPusher) *PaymentCallbackHandler {
	h.notifier.setPusher(p)
	return h
}

// WithPush mengaktifkan push notification (FCM). Panggil setelah WithHub.
func (h *MidtransHandler) WithPush(p *OrderPusher) *MidtransHandler {
	h.notifier.setPusher(p)
	return h
}

// Stream: GET /api/orders/stream  (customer / aplikasi mobile)
// Mengirim "event: orders" setiap kali pesanan milik customer berubah; klien lalu memuat ulang
// GET /api/orders. 404 bila siaran realtime tidak diaktifkan.
func (h *MobileHandler) Stream(w http.ResponseWriter, r *http.Request) {
	if h.hub == nil {
		respondNotFound(w, "Endpoint tidak ditemukan")
		return
	}
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		httpx.Error(w, http.StatusUnauthorized, "Tidak terautentikasi")
		return
	}
	sub, cancel := h.hub.Subscribe(user.ID)
	defer cancel()
	serveEvents(w, r, sub)
}

// Stream: GET /api/kasir/pesanan/stream  (web kasir)
// Mengirim "event: orders" setiap kali ada pesanan baru atau pesanan berubah (dari mobile,
// kasir lain, maupun pembayaran QRIS); klien lalu memuat ulang GET /api/kasir/pesanan.
func (h *KasirOrderHandler) Stream(w http.ResponseWriter, r *http.Request) {
	if h.hub == nil {
		respondNotFound(w, "Endpoint tidak ditemukan")
		return
	}
	sub, cancel := h.hub.SubscribeStaff()
	defer cancel()
	serveEvents(w, r, sub)
}

// serveEvents menulis aliran text/event-stream sampai klien terputus.
// sub sudah terdaftar sebelum pesan pembuka ditulis, jadi tidak ada perubahan yang terlewat.
func serveEvents(w http.ResponseWriter, r *http.Request, sub <-chan struct{}) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		httpx.Error(w, http.StatusInternalServerError, "Streaming tidak didukung")
		return
	}

	// Server memakai ReadTimeout/WriteTimeout 30 detik; keduanya akan memutus koneksi panjang ini.
	// Matikan khusus untuk stream (error diabaikan bila ResponseWriter tidak mendukung).
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Time{})
	_ = rc.SetWriteDeadline(time.Time{})

	hd := w.Header()
	hd.Set("Content-Type", "text/event-stream")
	hd.Set("Cache-Control", "no-cache, no-transform")
	hd.Set("Connection", "keep-alive")
	hd.Set("X-Accel-Buffering", "no") // cegah nginx/proxy menahan data
	w.WriteHeader(http.StatusOK)

	if _, err := fmt.Fprint(w, "retry: 3000\n: terhubung\n\n"); err != nil {
		return
	}
	flusher.Flush()

	ping := time.NewTicker(sseHeartbeat)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-sub:
			if _, err := fmt.Fprint(w, "event: orders\ndata: {}\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ping.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

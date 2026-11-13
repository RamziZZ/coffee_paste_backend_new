package handler

import (
	"context"
	"log"
	"sync"
	"time"

	"coffeepaste/internal/push"
	"coffeepaste/internal/store"
)

const maxTrackedOrders = 5000

type orderState struct {
	status string
	paid   bool
}

// OrderPusher mengirim push notification ke pemilik pesanan saat status atau pembayaran berubah.
// Satu instance dipakai bersama semua handler supaya perubahan yang sama tidak dikirim dua kali.
type OrderPusher struct {
	svc    *push.Service
	orders *store.OrderStore

	mu   sync.Mutex
	seen map[int64]orderState
}

// NewOrderPusher mengembalikan nil bila push tidak aktif.
func NewOrderPusher(svc *push.Service, orders *store.OrderStore) *OrderPusher {
	if svc == nil || orders == nil {
		return nil
	}
	return &OrderPusher{svc: svc, orders: orders, seen: make(map[int64]orderState)}
}

// Changed dipanggil setelah pesanan berubah. Pengiriman berjalan di goroutine sendiri.
func (p *OrderPusher) Changed(orderID, owner int64) {
	if p == nil || owner <= 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		o, err := p.orders.Get(ctx, orderID)
		if err != nil {
			log.Printf("push: gagal memuat pesanan %d: %v", orderID, err)
			return
		}
		if msg, ok := p.next(orderID, o); ok {
			p.svc.NotifyUser(ctx, owner, msg)
		}
	}()
}

// next membandingkan kondisi pesanan dengan yang terakhir dilihat; ok=false bila tidak ada yang perlu dikirim.
func (p *OrderPusher) next(id int64, o *store.Order) (push.Message, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	prev, known := p.seen[id]
	if len(p.seen) >= maxTrackedOrders {
		p.seen = make(map[int64]orderState)
	}
	p.seen[id] = orderState{status: o.Status, paid: o.Paid}

	statusChanged := (known && prev.status != o.Status) || (!known && o.Status != "menunggu")
	if statusChanged {
		if m, ok := statusMessage(o); ok {
			return m, true
		}
	}
	if o.Paid && (!known || !prev.paid) {
		return paymentMessage(o), true
	}
	return push.Message{}, false
}

func statusMessage(o *store.Order) (push.Message, bool) {
	var title, body string
	switch o.Status {
	case "proses":
		title, body = "Pesanan sedang dibuat", "Pesanan "+o.ID+" sedang disiapkan."
	case "siap":
		title, body = "Pesanan siap", "Pesanan "+o.ID+" sudah siap."
	case "selesai":
		title, body = "Pesanan selesai", "Terima kasih! Pesanan "+o.ID+" sudah selesai."
	case "dibatalkan":
		title, body = "Pesanan dibatalkan", "Pesanan "+o.ID+" dibatalkan."
		if o.CancelReason != nil && *o.CancelReason != "" {
			body += " Alasan: " + *o.CancelReason
		}
	default:
		return push.Message{}, false
	}
	return push.Message{
		Title: title,
		Body:  body,
		Data:  map[string]string{"type": "order", "orderId": o.ID, "status": o.Status},
	}, true
}

func paymentMessage(o *store.Order) push.Message {
	return push.Message{
		Title: "Pembayaran berhasil",
		Body:  "Pembayaran pesanan " + o.ID + " sudah diterima.",
		Data:  map[string]string{"type": "payment", "orderId": o.ID, "status": o.Status},
	}
}

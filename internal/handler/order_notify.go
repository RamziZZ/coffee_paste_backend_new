package handler

import (
	"context"
	"log"

	"coffeepaste/internal/realtime"
	"coffeepaste/internal/store"
)

// orderNotifier menyiarkan "pesanan berubah" ke semua kasir yang terhubung dan ke customer pemilik pesanan.
type orderNotifier struct {
	orders *store.OrderStore
	hub    *realtime.Hub
	pusher *OrderPusher // push notification (FCM); nil = nonaktif
}

func newOrderNotifier(orders *store.OrderStore, hub *realtime.Hub) *orderNotifier {
	if orders == nil || hub == nil {
		return nil
	}
	return &orderNotifier{orders: orders, hub: hub}
}

// changed dipanggil setelah pesanan dibuat atau berubah. Aman dipanggil pada notifier nil.
func (n *orderNotifier) changed(ctx context.Context, orderID int64) {
	if n == nil {
		return
	}
	// Kasir selalu diberi tahu, termasuk untuk pesanan tanpa akun customer.
	n.hub.NotifyStaff()

	owner, err := n.orders.CustomerOf(ctx, orderID)
	if err != nil {
		log.Printf("realtime: gagal mencari pemilik pesanan %d: %v", orderID, err)
		return
	}
	n.hub.Notify(owner)
	n.pusher.Changed(orderID, owner)
}

// setPusher mengaktifkan push notification. Aman dipanggil pada notifier nil.
func (n *orderNotifier) setPusher(p *OrderPusher) {
	if n != nil {
		n.pusher = p
	}
}

// changedByCode sama seperti changed, tetapi menerima kode pesanan seperti "#CP00001".
func (n *orderNotifier) changedByCode(ctx context.Context, code string) {
	if n == nil {
		return
	}
	id, ok := parseOrderCode(code)
	if !ok {
		n.hub.NotifyStaff()
		return
	}
	n.changed(ctx, id)
}

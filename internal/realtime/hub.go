
package realtime

import "sync"

// Hub menyimpan pelanggan per customer dan pelanggan staf (kasir).
type Hub struct {
	mu    sync.Mutex
	subs  map[int64]map[chan struct{}]struct{}
	staff map[chan struct{}]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subs:  make(map[int64]map[chan struct{}]struct{}),
		staff: make(map[chan struct{}]struct{}),
	}
}

// Subscribe mendaftarkan satu koneksi untuk customerID. Panggil cancel saat koneksi berakhir.
// Channel berkapasitas 1: banyak perubahan beruntun digabung menjadi satu sinyal.
func (h *Hub) Subscribe(customerID int64) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	if h == nil {
		return ch, func() {}
	}

	h.mu.Lock()
	if h.subs[customerID] == nil {
		h.subs[customerID] = make(map[chan struct{}]struct{})
	}
	h.subs[customerID][ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if set := h.subs[customerID]; set != nil {
			delete(set, ch)
			if len(set) == 0 {
				delete(h.subs, customerID)
			}
		}
	}
}

// Notify memberi sinyal ke semua koneksi milik customerID. Tidak pernah memblokir.
func (h *Hub) Notify(customerID int64) {
	if h == nil || customerID <= 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subs[customerID] {
		select {
		case ch <- struct{}{}:
		default: // sinyal sebelumnya belum dibaca; cukup satu
		}
	}
}

// SubscribeStaff mendaftarkan satu koneksi kasir. Kasir menerima sinyal untuk SEMUA perubahan pesanan.
// Panggil cancel saat koneksi berakhir.
func (h *Hub) SubscribeStaff() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	if h == nil {
		return ch, func() {}
	}

	h.mu.Lock()
	h.staff[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		delete(h.staff, ch)
	}
}

// NotifyStaff memberi sinyal ke semua koneksi kasir. Tidak pernah memblokir.
func (h *Hub) NotifyStaff() {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.staff {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

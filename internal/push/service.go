package push

import (
	"context"
	"errors"
	"log"
)

// TokenStore = penyimpanan token perangkat milik user.
type TokenStore interface {
	DeviceTokens(ctx context.Context, userID int64) ([]string, error)
	DeleteDeviceToken(ctx context.Context, token string) error
}

type Service struct {
	sender *Sender
	tokens TokenStore
}

// NewService mengembalikan nil bila sender/tokens kosong; semua method aman dipanggil pada nil.
func NewService(sender *Sender, tokens TokenStore) *Service {
	if sender == nil || tokens == nil {
		return nil
	}
	return &Service{sender: sender, tokens: tokens}
}

// NotifyUser mengirim pesan ke semua perangkat milik userID. Token mati dihapus otomatis.
func (s *Service) NotifyUser(ctx context.Context, userID int64, m Message) {
	if s == nil || userID <= 0 {
		return
	}
	tokens, err := s.tokens.DeviceTokens(ctx, userID)
	if err != nil {
		log.Printf("push: gagal membaca token user %d: %v", userID, err)
		return
	}
	for _, t := range tokens {
		err := s.sender.Send(ctx, t, m)
		switch {
		case errors.Is(err, ErrUnregistered):
			if derr := s.tokens.DeleteDeviceToken(ctx, t); derr != nil {
				log.Printf("push: gagal menghapus token mati: %v", derr)
			}
		case err != nil:
			log.Printf("push: gagal kirim ke user %d: %v", userID, err)
		}
	}
}

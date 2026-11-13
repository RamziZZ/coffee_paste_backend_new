package googleauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	certsURL     = "https://www.googleapis.com/oauth2/v3/certs"
	keyCacheTTL  = time.Hour
	minRefetch   = time.Minute // batas refetch agar kid ngawur tidak membanjiri Google
	maxCertsBody = 1 << 20
)

// Claims adalah data akun Google yang sudah terverifikasi.
type Claims struct {
	Subject       string // "sub": ID akun Google yang stabil (jangan pakai email sebagai ID)
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// Verifier memverifikasi ID token. Dibuat sebagai interface supaya handler mudah diuji.
type Verifier interface {
	Verify(ctx context.Context, idToken string) (*Claims, error)
}

// JWKSVerifier memverifikasi tanda tangan, issuer, audience, dan masa berlaku token.
type JWKSVerifier struct {
	audiences map[string]bool
	url       string
	client    *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

// New membuat verifier. audiences = daftar OAuth client ID yang boleh menjadi "aud" token
// (untuk aplikasi Flutter: Web client ID yang dipakai sebagai serverClientId).
func New(audiences []string) *JWKSVerifier {
	return newWithURL(audiences, certsURL, &http.Client{Timeout: 10 * time.Second})
}

func newWithURL(audiences []string, url string, client *http.Client) *JWKSVerifier {
	set := make(map[string]bool, len(audiences))
	for _, a := range audiences {
		if a = strings.TrimSpace(a); a != "" {
			set[a] = true
		}
	}
	return &JWKSVerifier{audiences: set, url: url, client: client}
}

type boolish bool

// UnmarshalJSON menerima true/false maupun "true"/"false".
func (b *boolish) UnmarshalJSON(data []byte) error {
	switch strings.Trim(string(data), `"`) {
	case "true":
		*b = true
	default:
		*b = false
	}
	return nil
}

type tokenClaims struct {
	jwt.RegisteredClaims
	Email         string  `json:"email"`
	EmailVerified boolish `json:"email_verified"`
	Name          string  `json:"name"`
	Picture       string  `json:"picture"`
}

var ErrInvalidToken = errors.New("ID token Google tidak valid")

// Verify mengembalikan klaim bila token sah. Semua kegagalan validasi dibungkus ErrInvalidToken.
func (v *JWKSVerifier) Verify(ctx context.Context, idToken string) (*Claims, error) {
	if len(v.audiences) == 0 {
		return nil, errors.New("googleauth: tidak ada client ID yang dikonfigurasi")
	}
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, fmt.Errorf("%w: kosong", ErrInvalidToken)
	}

	var tc tokenClaims
	_, err := jwt.ParseWithClaims(idToken, &tc, func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("kid tidak ada")
		}
		return v.key(ctx, kid)
	},
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if tc.Issuer != "https://accounts.google.com" && tc.Issuer != "accounts.google.com" {
		return nil, fmt.Errorf("%w: issuer %q", ErrInvalidToken, tc.Issuer)
	}
	okAud := false
	for _, a := range tc.Audience {
		if v.audiences[a] {
			okAud = true
			break
		}
	}
	if !okAud {
		return nil, fmt.Errorf("%w: audience tidak cocok", ErrInvalidToken)
	}
	if tc.Subject == "" {
		return nil, fmt.Errorf("%w: sub kosong", ErrInvalidToken)
	}

	return &Claims{
		Subject:       tc.Subject,
		Email:         tc.Email,
		EmailVerified: bool(tc.EmailVerified),
		Name:          tc.Name,
		Picture:       tc.Picture,
	}, nil
}

// key mengambil kunci publik untuk kid; memuat ulang JWKS bila cache basi atau kid belum dikenal.
func (v *JWKSVerifier) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	if k, ok := v.keys[kid]; ok && now.Sub(v.fetchedAt) < keyCacheTTL {
		return k, nil
	}
	if v.keys == nil || now.Sub(v.fetchedAt) >= minRefetch {
		if err := v.refresh(ctx); err != nil {
			if k, ok := v.keys[kid]; ok { // pakai cache lama bila Google sedang tidak terjangkau
				return k, nil
			}
			return nil, err
		}
	}
	k, ok := v.keys[kid]
	if !ok {
		return nil, errors.New("kunci penandatangan tidak dikenal")
	}
	return k, nil
}

func (v *JWKSVerifier) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.url, nil)
	if err != nil {
		return err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("ambil kunci Google: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ambil kunci Google: status %d", resp.StatusCode)
	}

	var set struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxCertsBody)).Decode(&set); err != nil {
		return fmt.Errorf("baca kunci Google: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		nb, err1 := base64.RawURLEncoding.DecodeString(k.N)
		eb, err2 := base64.RawURLEncoding.DecodeString(k.E)
		if err1 != nil || err2 != nil {
			continue
		}
		e := int(new(big.Int).SetBytes(eb).Int64())
		if e == 0 {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}
	}
	if len(keys) == 0 {
		return errors.New("kunci Google kosong")
	}
	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

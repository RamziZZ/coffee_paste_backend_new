package googleauth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testAud = "web-client-id.apps.googleusercontent.com"

func newTestVerifier(t *testing.T, kid string, pub *rsa.PublicKey) *JWKSVerifier {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": kid, "kty": "RSA", "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}}})
	}))
	t.Cleanup(srv.Close)
	return newWithURL([]string{testAud}, srv.URL, srv.Client())
}

func sign(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func goodClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss": "https://accounts.google.com", "aud": testAud, "sub": "1234567890",
		"email": "ayu@gmail.com", "email_verified": true, "name": "Ayu",
		"iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix(),
	}
}

func TestVerify(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := newTestVerifier(t, "k1", &key.PublicKey)

	t.Run("token sah", func(t *testing.T) {
		c, err := v.Verify(context.Background(), sign(t, key, "k1", goodClaims()))
		if err != nil {
			t.Fatal(err)
		}
		if c.Subject != "1234567890" || c.Email != "ayu@gmail.com" || !c.EmailVerified || c.Name != "Ayu" {
			t.Errorf("klaim salah: %+v", c)
		}
	})

	cases := map[string]func(jwt.MapClaims){
		"audience salah":   func(c jwt.MapClaims) { c["aud"] = "lain" },
		"issuer salah":     func(c jwt.MapClaims) { c["iss"] = "https://evil.example" },
		"kedaluwarsa":      func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() },
		"tanpa exp":        func(c jwt.MapClaims) { delete(c, "exp") },
		"sub kosong":       func(c jwt.MapClaims) { c["sub"] = "" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := goodClaims()
			mutate(c)
			if _, err := v.Verify(context.Background(), sign(t, key, "k1", c)); !errors.Is(err, ErrInvalidToken) {
				t.Errorf("err = %v, want ErrInvalidToken", err)
			}
		})
	}

	t.Run("ditandatangani kunci lain", func(t *testing.T) {
		if _, err := v.Verify(context.Background(), sign(t, other, "k1", goodClaims())); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("err = %v, want ErrInvalidToken", err)
		}
	})
	t.Run("kid tidak dikenal", func(t *testing.T) {
		if _, err := v.Verify(context.Background(), sign(t, key, "zzz", goodClaims())); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("err = %v, want ErrInvalidToken", err)
		}
	})
	t.Run("bukan jwt", func(t *testing.T) {
		if _, err := v.Verify(context.Background(), "abc"); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("err = %v, want ErrInvalidToken", err)
		}
	})
	t.Run("alg none ditolak", func(t *testing.T) {
		tok := jwt.NewWithClaims(jwt.SigningMethodNone, goodClaims())
		tok.Header["kid"] = "k1"
		s, _ := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if _, err := v.Verify(context.Background(), s); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("err = %v, want ErrInvalidToken", err)
		}
	})
}

func TestEmailVerifiedAsString(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := newTestVerifier(t, "k1", &key.PublicKey)
	c := goodClaims()
	c["email_verified"] = "true"
	got, err := v.Verify(context.Background(), sign(t, key, "k1", c))
	if err != nil || !got.EmailVerified {
		t.Fatalf("got %+v, err %v", got, err)
	}
}

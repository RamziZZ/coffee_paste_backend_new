package push

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// ChannelID harus sama dengan channel yang dibuat aplikasi Android.
	ChannelID = "orders"

	fcmScope        = "https://www.googleapis.com/auth/firebase.messaging"
	defaultTokenURI = "https://oauth2.googleapis.com/token"
	defaultAPIBase  = "https://fcm.googleapis.com"
)

// ErrUnregistered: token perangkat sudah tidak berlaku (aplikasi dihapus / token diganti).
var ErrUnregistered = errors.New("push: token perangkat tidak terdaftar")

type Message struct {
	Title string
	Body  string
	Data  map[string]string
}

type credentials struct {
	ProjectID   string `json:"project_id"`
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
	TokenURI    string `json:"token_uri"`
}

// Sender mengirim pesan lewat FCM HTTP v1 memakai service account Firebase.
type Sender struct {
	projectID string
	email     string
	key       *rsa.PrivateKey
	tokenURI  string
	apiBase   string
	client    *http.Client

	mu      sync.Mutex
	access  string
	expires time.Time
}

// ParseCredentials membaca JSON service account (teks JSON langsung atau base64-nya).
func ParseCredentials(raw string) (*Sender, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") {
		dec, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return nil, errors.New("kredensial Firebase bukan JSON maupun base64")
		}
		raw = string(dec)
	}

	var c credentials
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("JSON kredensial Firebase rusak: %w", err)
	}
	if c.ProjectID == "" || c.ClientEmail == "" || c.PrivateKey == "" {
		return nil, errors.New("kredensial Firebase tidak lengkap (project_id, client_email, private_key)")
	}

	// Beberapa hosting menyimpan newline kunci sebagai teks "\n".
	pem := strings.ReplaceAll(c.PrivateKey, `\n`, "\n")
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		return nil, fmt.Errorf("private_key Firebase tidak valid: %w", err)
	}
	if c.TokenURI == "" {
		c.TokenURI = defaultTokenURI
	}

	return &Sender{
		projectID: c.ProjectID,
		email:     c.ClientEmail,
		key:       key,
		tokenURI:  c.TokenURI,
		apiBase:   defaultAPIBase,
		client:    &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (s *Sender) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.access != "" && time.Now().Before(s.expires.Add(-time.Minute)) {
		return s.access, nil
	}

	now := time.Now()
	assertion, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":   s.email,
		"scope": fcmScope,
		"aud":   s.tokenURI,
		"iat":   now.Unix(),
		"exp":   now.Add(time.Hour).Unix(),
	}).SignedString(s.key)
	if err != nil {
		return "", err
	}

	form := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {assertion},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token OAuth ditolak (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.AccessToken == "" {
		return "", errors.New("respons token OAuth tidak valid")
	}
	s.access = out.AccessToken
	s.expires = now.Add(time.Duration(out.ExpiresIn) * time.Second)
	return s.access, nil
}

func (s *Sender) resetAccess() {
	s.mu.Lock()
	s.access = ""
	s.mu.Unlock()
}

// Send mengirim satu pesan ke satu token perangkat.
func (s *Sender) Send(ctx context.Context, deviceToken string, m Message) error {
	at, err := s.accessToken(ctx)
	if err != nil {
		return err
	}

	msg := map[string]any{
		"token":        deviceToken,
		"notification": map[string]string{"title": m.Title, "body": m.Body},
		// Prioritas tinggi supaya tetap tampil saat HP idle/aplikasi tertutup.
		"android": map[string]any{
			"priority":     "HIGH",
			"notification": map[string]string{"channel_id": ChannelID, "sound": "default"},
		},
		"apns": map[string]any{
			"headers": map[string]string{"apns-priority": "10"},
			"payload": map[string]any{"aps": map[string]any{"sound": "default"}},
		},
	}
	if len(m.Data) > 0 {
		msg["data"] = m.Data
	}
	payload, err := json.Marshal(map[string]any{"message": msg})
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/v1/projects/%s/messages:send", s.apiBase, s.projectID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+at)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusOK {
		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		s.resetAccess()
	}

	var e struct {
		Error struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Details []struct {
				ErrorCode string `json:"errorCode"`
			} `json:"details"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &e)
	if resp.StatusCode == http.StatusNotFound {
		return ErrUnregistered
	}
	for _, d := range e.Error.Details {
		if d.ErrorCode == "UNREGISTERED" {
			return ErrUnregistered
		}
	}
	return fmt.Errorf("FCM menolak pesan (%d %s): %s", resp.StatusCode, e.Error.Status, e.Error.Message)
}

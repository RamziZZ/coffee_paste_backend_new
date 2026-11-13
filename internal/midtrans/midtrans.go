// Package midtrans adalah klien kecil untuk Midtrans Core API (QRIS) dan verifikasi
// signature notifikasi. Tanpa dependensi tambahan.
package midtrans

import (
	"bytes"
	"context"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	sandboxURL    = "https://api.sandbox.midtrans.com"
	productionURL = "https://api.midtrans.com"
	maxBody       = 1 << 20
)

type Client struct {
	serverKey string
	baseURL   string
	notifyURL string // opsional: menimpa Payment Notification URL di dashboard
	http      *http.Client
}

// New membuat klien. production=false memakai sandbox. notifyURL boleh kosong
// (pakai Payment Notification URL yang diisi di dashboard Midtrans).
func New(serverKey string, production bool, notifyURL string) *Client {
	base := sandboxURL
	if production {
		base = productionURL
	}
	return &Client{
		serverKey: serverKey,
		baseURL:   base,
		notifyURL: strings.TrimSpace(notifyURL),
		http:      &http.Client{Timeout: 10 * time.Second},
	}
}

// QRISCharge adalah hasil pembuatan tagihan QRIS.
type QRISCharge struct {
	TransactionID string
	QRString      string // isi QR (dirender aplikasi mobile)
	QRURL         string // URL gambar QR dari Midtrans (untuk simulator/debug)
	ExpiryTime    string
}

type chargeResponse struct {
	StatusCode    string `json:"status_code"`
	StatusMessage string `json:"status_message"`
	TransactionID string `json:"transaction_id"`
	QRString      string `json:"qr_string"`
	ExpiryTime    string `json:"expiry_time"`
	Actions       []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"actions"`
}

// ChargeQRIS membuat tagihan QRIS. orderID harus unik per tagihan di Midtrans.
func (c *Client) ChargeQRIS(ctx context.Context, orderID string, amount int64, expiryMinutes int) (*QRISCharge, error) {
	if amount <= 0 {
		return nil, errors.New("midtrans: nominal harus lebih dari 0")
	}
	payload := map[string]any{
		"payment_type": "qris",
		"transaction_details": map[string]any{
			"order_id":     orderID,
			"gross_amount": amount,
		},
	}
	if expiryMinutes > 0 {
		payload["custom_expiry"] = map[string]any{
			"expiry_duration": expiryMinutes,
			"unit":            "minute",
		}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v2/charge", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))
	if c.notifyURL != "" {
		req.Header.Set("X-Override-Notification", c.notifyURL)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("midtrans: baca respons: %w", err)
	}

	var cr chargeResponse
	if err := json.Unmarshal(raw, &cr); err != nil {
		return nil, fmt.Errorf("midtrans: respons bukan JSON (HTTP %d)", resp.StatusCode)
	}
	// Midtrans membalas HTTP 200 dengan status_code di body (201 = tagihan dibuat).
	if cr.StatusCode != "200" && cr.StatusCode != "201" {
		snippet := strings.TrimSpace(string(raw))
		if len(snippet) > 400 {
			snippet = snippet[:400] + "..."
		}
		return nil, fmt.Errorf("midtrans: HTTP %d, status_code=%q, pesan=%q, body=%s (url=%s)",
			resp.StatusCode, cr.StatusCode, cr.StatusMessage, snippet, c.baseURL)
	}
	if cr.QRString == "" {
		return nil, errors.New("midtrans: respons tidak memuat qr_string")
	}

	out := &QRISCharge{TransactionID: cr.TransactionID, QRString: cr.QRString, ExpiryTime: cr.ExpiryTime}
	for _, a := range cr.Actions {
		if a.Name == "generate-qr-code" {
			out.QRURL = a.URL
			break
		}
	}
	return out, nil
}

// Signature menghitung signature_key notifikasi:
// SHA512(order_id + status_code + gross_amount + serverKey), heksadesimal huruf kecil.
func Signature(orderID, statusCode, grossAmount, serverKey string) string {
	sum := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	return hex.EncodeToString(sum[:])
}

// VerifySignature memeriksa signature_key notifikasi dengan perbandingan waktu-konstan.
func VerifySignature(orderID, statusCode, grossAmount, signature, serverKey string) bool {
	if serverKey == "" || signature == "" {
		return false
	}
	want := Signature(orderID, statusCode, grossAmount, serverKey)
	return subtle.ConstantTimeCompare([]byte(want), []byte(strings.ToLower(strings.TrimSpace(signature)))) == 1
}

// ErrTransactionNotFound dikembalikan Status bila Midtrans belum/tidak mengenal order_id itu.
var ErrTransactionNotFound = errors.New("midtrans: transaksi tidak ditemukan")

// flexString menerima nilai JSON berupa string atau angka (Midtrans kadang mengirim angka).
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*f = ""
		return nil
	}
	if strings.HasPrefix(s, `"`) {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*f = flexString(v)
		return nil
	}
	*f = flexString(s)
	return nil
}

// TxStatus adalah status transaksi yang dibaca langsung dari Midtrans (GET /v2/{order_id}/status).
type TxStatus struct {
	OrderID           string
	StatusCode        string // "200" = sukses dibaca, "404" = belum ada
	TransactionStatus string // pending | settlement | capture | expire | cancel | deny | failure
	FraudStatus       string
	GrossAmount       string // mis. "55500.00"
}

// Status menanyakan status transaksi ke Midtrans. Dipakai sebagai cadangan notifikasi:
// notifikasi (webhook) tidak akan sampai bila server tidak bisa dijangkau dari internet
// (mis. dijalankan di localhost) atau Payment Notification URL belum diisi di dashboard.
func (c *Client) Status(ctx context.Context, orderID string) (*TxStatus, error) {
	if strings.TrimSpace(orderID) == "" {
		return nil, errors.New("midtrans: order_id kosong")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v2/"+url.PathEscape(orderID)+"/status", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.serverKey+":")))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, fmt.Errorf("midtrans: baca respons: %w", err)
	}

	var sr struct {
		StatusCode        flexString `json:"status_code"`
		StatusMessage     string     `json:"status_message"`
		OrderID           flexString `json:"order_id"`
		TransactionStatus flexString `json:"transaction_status"`
		FraudStatus       flexString `json:"fraud_status"`
		GrossAmount       flexString `json:"gross_amount"`
	}
	if err := json.Unmarshal(raw, &sr); err != nil {
		return nil, fmt.Errorf("midtrans: respons bukan JSON (HTTP %d)", resp.StatusCode)
	}
	switch string(sr.StatusCode) {
	case "404":
		return nil, ErrTransactionNotFound
	case "401":
		return nil, fmt.Errorf("midtrans: HTTP %d, kunci server ditolak (401)", resp.StatusCode)
	}
	return &TxStatus{
		OrderID:           string(sr.OrderID),
		StatusCode:        string(sr.StatusCode),
		TransactionStatus: strings.ToLower(string(sr.TransactionStatus)),
		FraudStatus:       strings.ToLower(string(sr.FraudStatus)),
		GrossAmount:       string(sr.GrossAmount),
	}, nil
}

package midtrans

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") == "" {
			t.Errorf("request salah: %s auth=%q", r.Method, r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/v2/CP00001-1/status":
			w.Write([]byte(`{"status_code":"200","order_id":"CP00001-1","transaction_status":"settlement","gross_amount":"55500.00"}`))
		case "/v2/CP00002-1/status":
			w.Write([]byte(`{"status_code":404,"status_message":"Transaction doesn't exist."}`))
		}
	}))
	defer srv.Close()
	c := New("SB-test", false, "")
	c.baseURL = srv.URL

	st, err := c.Status(context.Background(), "CP00001-1")
	if err != nil || st.TransactionStatus != "settlement" || st.GrossAmount != "55500.00" {
		t.Fatalf("got %+v err=%v", st, err)
	}
	if _, err := c.Status(context.Background(), "CP00002-1"); !errors.Is(err, ErrTransactionNotFound) {
		t.Fatalf("want ErrTransactionNotFound, got %v", err)
	}
}

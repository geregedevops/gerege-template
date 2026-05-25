// Gerege Template Version 27.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

package verify

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Тест бүр httptest.Server-ээр алсын /send, /check endpoint-уудыг хуурамчаар
// гаргадаг тул бид төлөв засах ажилгүйгээр HTTP wire-формат, header-уудыг
// шалгаж чадна. Жинхэнэ verify.gecloud.mn-ийг хэзээ ч цохиогүй.

func TestSend_HappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/send" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "clk_live_test" {
			t.Errorf("missing/incorrect X-API-Key: %q", r.Header.Get("X-API-Key"))
		}
		body, _ := io.ReadAll(r.Body)
		got := string(body)
		if !strings.Contains(got, `"to":"99097189"`) || !strings.Contains(got, `"channel":"sms"`) {
			t.Errorf("unexpected request body: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_id":"clv_abc123"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "clk_live_test", ChannelSMS)
	rid, err := c.Send(context.Background(), "99097189")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rid != "clv_abc123" {
		t.Fatalf("expected request_id clv_abc123, got %s", rid)
	}
}

func TestCheck_Approved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/check" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		got := string(body)
		if !strings.Contains(got, `"request_id":"clv_abc123"`) || !strings.Contains(got, `"code":"482916"`) {
			t.Errorf("unexpected request body: %s", got)
		}
		_, _ = w.Write([]byte(`{"status":"approved"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "clk_live_test", ChannelEmail)
	if err := c.Check(context.Background(), "clv_abc123", "482916"); err != nil {
		t.Fatalf("expected approved, got %v", err)
	}
}

func TestCheck_NotApproved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"pending"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "clk_live_test", ChannelEmail)
	err := c.Check(context.Background(), "clv_abc123", "000000")
	if !errors.Is(err, ErrNotApproved) {
		t.Fatalf("expected ErrNotApproved, got %v", err)
	}
}

func TestSend_NonOK_BubblesServerMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"insufficient balance"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "clk_live_test", ChannelEmail)
	_, err := c.Send(context.Background(), "patrick@example.com")
	if err == nil || !strings.Contains(err.Error(), "insufficient balance") {
		t.Fatalf("expected error to surface server message, got %v", err)
	}
}

func TestSend_MissingAPIKey(t *testing.T) {
	c := NewClient("https://example.invalid", "", ChannelEmail)
	if _, err := c.Send(context.Background(), "x@example.com"); err == nil {
		t.Fatalf("expected missing api key error")
	}
}

func TestNewClient_TrimsTrailingSlash_Defaults(t *testing.T) {
	c := NewClient("https://example.invalid/", "k", "")
	if c.baseURL != "https://example.invalid" {
		t.Errorf("expected trailing slash trimmed, got %q", c.baseURL)
	}
	if c.channel != ChannelEmail {
		t.Errorf("expected default channel 'email', got %q", c.channel)
	}
}

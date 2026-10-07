package events

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSenderForwardsEventAndReturnsUpstreamResponse(t *testing.T) {
	const payload = `{"event":"test"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			return
		}
		if r.Method != http.MethodPatch {
			t.Errorf("request method = %q, want PATCH", r.Method)
		}
		if got := string(body); got != payload {
			t.Errorf("request body = %q, want %q", got, payload)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		w.Header().Set("X-Upstream", "received")
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer server.Close()

	response, err := NewHTTPSender().Send(context.Background(), http.MethodPatch, server.URL, []byte(payload))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if response.StatusCode != http.StatusAccepted || string(response.Body) != `{"ok":true}` {
		t.Fatalf("Send() response = %#v", response)
	}
	if got := response.Header.Get("X-Upstream"); got != "received" {
		t.Fatalf("upstream header = %q, want received", got)
	}
}

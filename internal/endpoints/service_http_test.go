package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestHTTPCheckerRetriesUntilEndpointRespondsSuccessfully(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := NewHTTPChecker().Check(context.Background(), server.URL); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if got := attempts.Load(); got != 3 {
		t.Fatalf("Check() made %d requests, want 3", got)
	}
}

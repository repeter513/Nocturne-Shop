package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIPRateLimitBlocksAfterMax(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	l := NewIPRateLimit(2, time.Minute)
	l.nowFunc = func() time.Time { return fixed }

	var hits int
	h := l.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.1:1234"

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i+1, rr.Code)
		}
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status %d, want 429", rr.Code)
	}
	if hits != 2 {
		t.Fatalf("hits = %d, want 2", hits)
	}
}

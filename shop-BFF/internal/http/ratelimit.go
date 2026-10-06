// IP rate limiting for public auth REST routes (complements Envoy edge limits).
// Rate limit по IP для публичных auth REST (дополняет лимиты Envoy на edge).
package http

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ipWindow struct {
	count int
	reset time.Time
}

// IPRateLimit caps requests per client IP in a sliding window (X-Forwarded-For / X-Real-IP).
// IPRateLimit ограничивает запросы с IP клиента в окне (X-Forwarded-For / X-Real-IP).
type IPRateLimit struct {
	mu      sync.Mutex
	byIP    map[string]ipWindow
	max     int
	window  time.Duration
	nowFunc func() time.Time
}

// NewIPRateLimit builds a limiter; maxPerWindow defaults to 30 when <= 0.
// NewIPRateLimit создаёт лимитер; при maxPerWindow <= 0 используется 30.
func NewIPRateLimit(maxPerWindow int, window time.Duration) *IPRateLimit {
	if maxPerWindow <= 0 {
		maxPerWindow = 30
	}
	if window <= 0 {
		window = time.Minute
	}
	return &IPRateLimit{
		byIP:    make(map[string]ipWindow),
		max:     maxPerWindow,
		window:  window,
		nowFunc: time.Now,
	}
}

func (l *IPRateLimit) allow(ip string) bool {
	now := l.nowFunc()
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.byIP[ip]
	if !ok || now.After(w.reset) {
		l.byIP[ip] = ipWindow{count: 1, reset: now.Add(l.window)}
		return true
	}
	if w.count >= l.max {
		return false
	}
	w.count++
	l.byIP[ip] = w
	return true
}

func (l *IPRateLimit) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(clientIP(r)) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			xff = xff[:i]
		}
		xff = strings.TrimSpace(xff)
		if xff != "" {
			return xff
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

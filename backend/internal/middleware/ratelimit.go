package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter is a simple in-memory, per-IP token bucket limiter.
// Each IP gets `burst` requests immediately, refilled at `perMinute` per minute.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64 // tokens per second
	burst    float64
}

type visitor struct {
	tokens float64
	last   time.Time
}

func NewRateLimiter(perMinute, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(perMinute) / 60.0,
		burst:    float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[key]
	if !ok {
		rl.visitors[key] = &visitor{tokens: rl.burst - 1, last: now}
		return true
	}

	v.tokens += now.Sub(v.last).Seconds() * rl.rate
	if v.tokens > rl.burst {
		v.tokens = rl.burst
	}
	v.last = now

	if v.tokens < 1 {
		return false
	}
	v.tokens--
	return true
}

// cleanup forgets IPs that have been idle for a while so memory stays small.
func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.last) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) reject(w http.ResponseWriter) {
	EnableCors(w)
	w.Header().Set("Retry-After", "60")
	http.Error(w, "Too many requests, please slow down", http.StatusTooManyRequests)
}

// Limit applies the limiter to every request (except CORS preflight).
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions && !rl.allow(ClientIP(r)) {
			rl.reject(w)
			return
		}
		next(w, r)
	}
}

// LimitWrites applies the limiter only to requests that change data
// (POST, PUT, DELETE, ...). Reads are not counted.
func (rl *RateLimiter) LimitWrites(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if !rl.allow(ClientIP(r)) {
				rl.reject(w)
				return
			}
		}
		next(w, r)
	}
}

// ClientIP returns the visitor's IP address.
// In production the backend is only reachable through Caddy, which sets
// X-Forwarded-For to the real client IP (it ignores values sent by clients).
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.TrimSpace(strings.Split(xff, ",")[0]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

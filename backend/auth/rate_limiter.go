package auth

import (
	"net/http"
	"sync"
	"time"
)

// bucket tracks the token bucket state for a single key (IP address).
type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimiter is a simple in-memory per-IP token-bucket rate limiter.
// It is safe for concurrent use.
type RateLimiter struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	rate      float64 // tokens per second
	burst     int     // max tokens
	lastPrune time.Time
}

// NewRateLimiter creates a rate limiter that allows `rate` requests per second
// with a burst capacity of `burst` requests.
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		buckets:   make(map[string]*bucket),
		rate:      rate,
		burst:     burst,
		lastPrune: time.Now(),
	}
}

// Allow reports whether a request from the given key (typically an IP) should
// be permitted. Returns false if the rate limit has been exceeded.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Opportunistic pruning of stale entries every 5 minutes.
	if now.Sub(rl.lastPrune) > 5*time.Minute {
		for k, b := range rl.buckets {
			if now.Sub(b.lastCheck) > 10*time.Minute {
				delete(rl.buckets, k)
			}
		}
		rl.lastPrune = now
	}

	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{
			tokens:    float64(rl.burst) - 1,
			lastCheck: now,
		}
		return true
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastCheck = now

	if b.tokens < 1 {
		return false
	}

	b.tokens--
	return true
}

// RateLimitMiddleware returns HTTP middleware that enforces per-IP rate limiting.
// Requests that exceed the limit receive a 429 Too Many Requests response.
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			// chi's RealIP middleware sets X-Real-IP; prefer it when available.
			if realIP := r.Header.Get("X-Real-Ip"); realIP != "" {
				ip = realIP
			}

			if !limiter.Allow(ip) {
				w.Header().Set("Retry-After", "10")
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

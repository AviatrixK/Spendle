package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/YOUR_GITHUB_USERNAME/spendle/internal/httpx"
)

// clientBucket tracks request counts for a single client IP address.
type clientBucket struct {
	count     int
	resetTime time.Time
}

// IPRateLimiter provides thread-safe in-memory rate limiting based on client IP addresses.
//
// Go Concept - sync.Mutex for Concurrent Memory Protection:
// HTTP handlers in Go run concurrently across separate goroutines. When multiple requests
// from the same IP arrive at the same time, modifying a shared map (`clients`) without
// synchronization causes race conditions and fatal runtime crashes.
// `sync.Mutex` ensures that only one goroutine reads or writes the map at any instant.
type IPRateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]*clientBucket
}

// NewIPRateLimiter creates an in-memory rate limiter and starts a background garbage-collection
// goroutine to prevent memory leaks from inactive IPs over time.
func NewIPRateLimiter(limit int, window time.Duration) *IPRateLimiter {
	limiter := &IPRateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientBucket),
	}

	// Go Concept - Background Cleanup Goroutine:
	// A lightweight goroutine periodically cleans up stale IPs so memory usage remains bounded.
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			limiter.mu.Lock()
			now := time.Now()
			for ip, bucket := range limiter.clients {
				if now.After(bucket.resetTime) {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

// Limit wraps an http.Handler with rate limiting enforcement.
// If the client exceeds `limit` requests within `window`, HTTP 429 Too Many Requests is returned.
func (l *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		l.mu.Lock()
		now := time.Now()
		bucket, exists := l.clients[ip]

		if !exists || now.After(bucket.resetTime) {
			// Initialize or reset window for this IP
			l.clients[ip] = &clientBucket{
				count:     1,
				resetTime: now.Add(l.window),
			}
			l.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		if bucket.count < l.limit {
			bucket.count++
			l.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}

		// Rate limit reached
		retryAfterSeconds := int(time.Until(bucket.resetTime).Seconds()) + 1
		if retryAfterSeconds < 1 {
			retryAfterSeconds = 1
		}
		l.mu.Unlock()

		w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
		_ = httpx.WriteError(w, http.StatusTooManyRequests, "Too many requests. Please try again in a few moments.")
	})
}

// getClientIP resolves the real client IP address, properly inspecting proxy headers
// set by platforms like Render, Vercel, and Cloudflare.
func getClientIP(r *http.Request) string {
	// 1. X-Forwarded-For: Client IP is the first entry in comma-separated chain
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		first := strings.TrimSpace(parts[0])
		if first != "" {
			return first
		}
	}

	// 2. X-Real-IP: Set by reverse proxies (e.g., NGINX, Cloudflare)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. Fallback to direct TCP connection address
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}


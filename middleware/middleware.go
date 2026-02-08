package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"practice2/models"
)

const RequiredAPIKey = "secret12345"
const RequestsPerSecond = 10

func APIKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey != RequiredAPIKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(models.ErrorResponse{Error: "unauthorized"})
			return
		}

		next(w, r)
	}
}

type LoggingMiddleware struct {
	handler http.Handler
}

func NewLoggingMiddleware(handler http.Handler) *LoggingMiddleware {
	return &LoggingMiddleware{handler: handler}
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timestamp := time.Now().Format(time.RFC3339)
	method := r.Method
	path := r.URL.Path

	l.handler.ServeHTTP(w, r)

	println(timestamp, method, path, "request processed")
}

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientBucket
}

type clientBucket struct {
	tokens     int64
	lastRefill time.Time
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		clients: make(map[string]*clientBucket),
	}
}

func (rl *RateLimiter) allow(apiKey string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.clients[apiKey]

	if !exists {
		bucket = &clientBucket{
			tokens:     RequestsPerSecond - 1,
			lastRefill: now,
		}
		rl.clients[apiKey] = bucket
		return true
	}

	elapsed := now.Sub(bucket.lastRefill)
	seconds := int64(elapsed / time.Second)

	if seconds >= 1 {
		bucket.tokens = RequestsPerSecond
		bucket.lastRefill = bucket.lastRefill.Add(time.Duration(seconds) * time.Second)
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func RateLimitMiddleware(limiter *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-KEY")

			if !limiter.allow(apiKey) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(models.ErrorResponse{Error: "rate limit exceeded"})
				return
			}

			next(w, r)
		}
	}
}
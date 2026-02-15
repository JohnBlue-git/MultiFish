package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"multifish/utility"
)

// RateLimiter holds rate limiters for different clients
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
// rate: requests per second (e.g., 10 for 10 req/s)
// burst: maximum burst size (e.g., 20 for up to 20 requests at once)
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// getLimiter returns the rate limiter for a given key (usually IP address)
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimitMiddleware creates a Gin middleware for rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use client IP as the key
		key := c.ClientIP()
		limiter := rl.getLimiter(key)

		if !limiter.Allow() {
			log := utility.GetLogger()
			log.Warn().
				Str("ip", key).
				Str("path", c.Request.URL.Path).
				Msg("Rate limit exceeded")

			utility.RedfishError(c, http.StatusTooManyRequests, 
				"Rate limit exceeded. Please try again later.", 
				"RateLimitExceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupStaleEntries removes rate limiters that haven't been used recently
// This should be called periodically to prevent memory leaks
func (rl *RateLimiter) CleanupStaleEntries(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Note: This is a simplified cleanup. In production, you might want to track
	// last access time for each limiter
	if len(rl.limiters) > 10000 {
		// Reset if too many entries (simple protection)
		rl.limiters = make(map[string]*rate.Limiter)
	}
}

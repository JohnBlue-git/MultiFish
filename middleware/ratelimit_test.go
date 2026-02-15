package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a rate limiter: 2 requests per second, burst of 3
	rl := NewRateLimiter(rate.Limit(2), 3)

	// Create a test router
	router := gin.New()
	router.Use(rl.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test 1: First 3 requests should succeed (burst)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345" // Same IP
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d failed: expected status 200, got %d", i+1, w.Code)
		}
	}

	// Test 2: 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected rate limit (429), got %d", w.Code)
	}

	// Test 3: Wait and try again (should succeed)
	time.Sleep(600 * time.Millisecond) // Wait for rate limit to reset
	req = httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("After waiting, expected status 200, got %d", w.Code)
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a rate limiter: 1 request per second, burst of 1
	rl := NewRateLimiter(rate.Limit(1), 1)

	router := gin.New()
	router.Use(rl.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Request from IP 1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First IP request failed: expected 200, got %d", w1.Code)
	}

	// Request from IP 2 (should succeed, different limiter)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second IP request failed: expected 200, got %d", w2.Code)
	}

	// Second request from IP 1 (should be rate limited)
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("Second request from IP1: expected 429, got %d", w3.Code)
	}
}

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name  string
		rate  rate.Limit
		burst int
	}{
		{
			name:  "low rate",
			rate:  rate.Limit(1),
			burst: 1,
		},
		{
			name:  "medium rate",
			rate:  rate.Limit(10),
			burst: 20,
		},
		{
			name:  "high rate",
			rate:  rate.Limit(100),
			burst: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rate, tt.burst)
			if rl == nil {
				t.Error("NewRateLimiter() returned nil")
			}
			if rl.rate != tt.rate {
				t.Errorf("rate = %v, want %v", rl.rate, tt.rate)
			}
			if rl.burst != tt.burst {
				t.Errorf("burst = %v, want %v", rl.burst, tt.burst)
			}
		})
	}
}

func TestCleanupStaleEntries(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 20)

	// Add many limiters
	for i := 0; i < 15000; i++ {
		key := string(rune(i))
		rl.getLimiter(key)
	}

	// Cleanup should reset when over 10000
	rl.CleanupStaleEntries(1 * time.Hour)

	if len(rl.limiters) > 10000 {
		t.Errorf("CleanupStaleEntries() did not clean up: got %d limiters", len(rl.limiters))
	}
}

package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"multifish/utility"
)

func init() {
	// Initialize logger for tests
	utility.InitLogger("error")
	gin.SetMode(gin.TestMode)
}

// TestAuthMiddleware_Disabled tests that middleware passes through when auth is disabled
func TestAuthMiddleware_Disabled(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: false,
		Mode:    "none",
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test without any credentials
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_BasicAuth_Success tests successful basic authentication
func TestAuthMiddleware_BasicAuth_Success(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "basic",
		BasicAuth: &BasicAuthConfig{
			Username: "admin",
			Password: "secret123",
		},
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with correct credentials
	req, _ := http.NewRequest("GET", "/test", nil)
	auth := base64.StdEncoding.EncodeToString([]byte("admin:secret123"))
	req.Header.Set("Authorization", "Basic "+auth)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_BasicAuth_Failure tests failed basic authentication
func TestAuthMiddleware_BasicAuth_Failure(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "basic",
		BasicAuth: &BasicAuthConfig{
			Username: "admin",
			Password: "secret123",
		},
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name     string
		username string
		password string
	}{
		{"wrong password", "admin", "wrongpass"},
		{"wrong username", "wronguser", "secret123"},
		{"both wrong", "wrong", "wrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			auth := base64.StdEncoding.EncodeToString([]byte(tt.username + ":" + tt.password))
			req.Header.Set("Authorization", "Basic "+auth)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestAuthMiddleware_BasicAuth_NoCredentials tests basic auth without credentials
func TestAuthMiddleware_BasicAuth_NoCredentials(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "basic",
		BasicAuth: &BasicAuthConfig{
			Username: "admin",
			Password: "secret123",
		},
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get("WWW-Authenticate"), "Basic realm")
}

// TestAuthMiddleware_TokenAuth_Success tests successful token authentication
func TestAuthMiddleware_TokenAuth_Success(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "token",
		TokenAuth: &TokenAuthConfig{
			Tokens: []string{
				"my-secret-token-123",
				"another-valid-token-456",
			},
		},
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name  string
		token string
	}{
		{"first token", "my-secret-token-123"},
		{"second token", "another-valid-token-456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// TestAuthMiddleware_TokenAuth_Failure tests failed token authentication
func TestAuthMiddleware_TokenAuth_Failure(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "token",
		TokenAuth: &TokenAuthConfig{
			Tokens: []string{"valid-token-123"},
		},
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"invalid token", "Bearer invalid-token"},
		{"no bearer prefix", "valid-token-123"},
		{"wrong prefix", "Basic valid-token-123"},
		{"no header", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestAuthMiddleware_ModeNone tests that mode "none" allows all requests
func TestAuthMiddleware_ModeNone(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true, // Even with enabled=true
		Mode:    "none",
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestAuthMiddleware_InvalidMode tests handling of invalid auth mode
func TestAuthMiddleware_InvalidMode(t *testing.T) {
	router := gin.New()
	
	authCfg := &AuthConfig{
		Enabled: true,
		Mode:    "invalid-mode",
	}
	
	router.Use(AuthMiddleware(authCfg))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

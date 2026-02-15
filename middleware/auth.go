package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"multifish/utility"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	Mode          string            `yaml:"mode" json:"mode"` // "basic", "token", or "none"
	BasicAuth     *BasicAuthConfig  `yaml:"basic_auth" json:"basic_auth"`
	TokenAuth     *TokenAuthConfig  `yaml:"token_auth" json:"token_auth"`
}

// BasicAuthConfig holds basic authentication configuration
type BasicAuthConfig struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// TokenAuthConfig holds token authentication configuration
type TokenAuthConfig struct {
	Tokens []string `yaml:"tokens" json:"tokens"` // List of valid tokens
}

// AuthMiddleware returns a Gin middleware for authentication
func AuthMiddleware(authCfg *AuthConfig) gin.HandlerFunc {
	log := utility.GetLogger()

	return func(c *gin.Context) {
		// Skip authentication if disabled
		if !authCfg.Enabled || authCfg.Mode == "none" {
			c.Next()
			return
		}

		// Handle different authentication modes
		switch strings.ToLower(authCfg.Mode) {
		case "basic":
			if !validateBasicAuth(c, authCfg.BasicAuth) {
				log.Warn().
					Str("ip", c.ClientIP()).
					Str("path", c.Request.URL.Path).
					Msg("Unauthorized access attempt (Basic Auth)")
				
				c.Header("WWW-Authenticate", `Basic realm="MultiFish API"`)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized",
					"message": "Valid Basic Authentication credentials required",
				})
				return
			}

		case "token":
			if !validateTokenAuth(c, authCfg.TokenAuth) {
				log.Warn().
					Str("ip", c.ClientIP()).
					Str("path", c.Request.URL.Path).
					Msg("Unauthorized access attempt (Token Auth)")
				
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Unauthorized",
					"message": "Valid Bearer token required",
				})
				return
			}

		default:
			log.Error().
				Str("mode", authCfg.Mode).
				Msg("Invalid authentication mode configured")
			
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
				"message": "Invalid authentication configuration",
			})
			return
		}

		c.Next()
	}
}

// validateBasicAuth validates Basic Authentication credentials
func validateBasicAuth(c *gin.Context, basicCfg *BasicAuthConfig) bool {
	if basicCfg == nil || basicCfg.Username == "" || basicCfg.Password == "" {
		return false
	}

	username, password, ok := c.Request.BasicAuth()
	if !ok {
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(basicCfg.Username)) == 1
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(basicCfg.Password)) == 1

	return usernameMatch && passwordMatch
}

// validateTokenAuth validates Bearer Token authentication
func validateTokenAuth(c *gin.Context, tokenCfg *TokenAuthConfig) bool {
	if tokenCfg == nil || len(tokenCfg.Tokens) == 0 {
		return false
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return false
	}

	// Extract token from "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return false
	}

	token := parts[1]
	
	// Check if token is in the list of valid tokens
	for _, validToken := range tokenCfg.Tokens {
		if subtle.ConstantTimeCompare([]byte(token), []byte(validToken)) == 1 {
			return true
		}
	}

	return false
}

// DefaultAuthConfig returns default authentication configuration (disabled)
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled: false,
		Mode:    "none",
		BasicAuth: &BasicAuthConfig{
			Username: "",
			Password: "",
		},
		TokenAuth: &TokenAuthConfig{
			Tokens: []string{},
		},
	}
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"multifish/middleware"
	"multifish/utility"
)

// Config holds all application configuration
type Config struct {
	Port              int                       `yaml:"port" json:"port"`
	LogLevel          string                    `yaml:"log_level" json:"log_level"`
	WorkerPoolSize    int                       `yaml:"worker_pool_size" json:"worker_pool_size"`
	LogsDir           string                    `yaml:"logs_dir" json:"logs_dir"`
	ShutdownTimeout   int                       `yaml:"shutdown_timeout" json:"shutdown_timeout"`       // Graceful shutdown timeout in seconds
	RateLimitRate     float64                   `yaml:"rate_limit_rate" json:"rate_limit_rate"`         // Requests per second
	RateLimitBurst    int                       `yaml:"rate_limit_burst" json:"rate_limit_burst"`       // Maximum burst size
	RateLimitEnabled  bool                      `yaml:"rate_limit_enabled" json:"rate_limit_enabled"`   // Enable/disable rate limiting
	Auth              *middleware.AuthConfig    `yaml:"auth" json:"auth"`                               // Authentication configuration
}

// DefaultConfig returns default configuration values
func DefaultConfig() *Config {
	return &Config{
		Port:             8080,
		LogLevel:         "info",
		WorkerPoolSize:   99,
		LogsDir:          "./logs",
		ShutdownTimeout:  30,    // 30 seconds for graceful shutdown
		RateLimitRate:    10.0,  // 10 requests per second
		RateLimitBurst:   20,    // Allow bursts up to 20 requests
		RateLimitEnabled: true,  // Rate limiting enabled by default
		Auth:             middleware.DefaultAuthConfig(), // Authentication disabled by default
	}
}

// LoadConfig loads configuration from environment variables and optional YAML file
// Priority: Environment Variables > YAML File > Defaults
func LoadConfig(configPath string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()
	log := utility.GetLogger()

	// Load from YAML file if provided
	if configPath != "" {
		if err := cfg.loadFromFile(configPath); err != nil {
			log.Warn().Msgf("Failed to load config file '%s': %v. Using defaults and environment variables.", configPath, err)
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables (highest priority)
	cfg.loadFromEnv()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Error().Msgf("Invalid configuration: %v", err)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file
func (c *Config) loadFromFile(path string) error {
	log := utility.GetLogger()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Msgf("Config file not found: %s", path)
			return fmt.Errorf("configuration file not found at path '%s'. Create a config file using 'cp config.example.yaml %s' or specify a different path", path, path)
		}
		log.Error().Msgf("Failed to read config file '%s': %v", path, err)
		return err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		log.Error().Msgf("Failed to parse YAML config file '%s': %v", path, err)
		return fmt.Errorf("failed to parse YAML configuration file '%s': %w. Check YAML syntax and indentation", path, err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func (c *Config) loadFromEnv() {
	// PORT
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}

	// LOG_LEVEL
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	// WORKER_POOL_SIZE
	if workerPoolSize := os.Getenv("WORKER_POOL_SIZE"); workerPoolSize != "" {
		if w, err := strconv.Atoi(workerPoolSize); err == nil {
			c.WorkerPoolSize = w
		}
	}

	// LOGS_DIR
	if logsDir := os.Getenv("LOGS_DIR"); logsDir != "" {
		c.LogsDir = logsDir
	}

	// SHUTDOWN_TIMEOUT
	if shutdownTimeout := os.Getenv("SHUTDOWN_TIMEOUT"); shutdownTimeout != "" {
		if s, err := strconv.Atoi(shutdownTimeout); err == nil {
			c.ShutdownTimeout = s
		}
	}

	// RATE_LIMIT_RATE
	if rateLimitRate := os.Getenv("RATE_LIMIT_RATE"); rateLimitRate != "" {
		if r, err := strconv.ParseFloat(rateLimitRate, 64); err == nil {
			c.RateLimitRate = r
		}
	}

	// RATE_LIMIT_BURST
	if rateLimitBurst := os.Getenv("RATE_LIMIT_BURST"); rateLimitBurst != "" {
		if b, err := strconv.Atoi(rateLimitBurst); err == nil {
			c.RateLimitBurst = b
		}
	}

	// RATE_LIMIT_ENABLED
	if rateLimitEnabled := os.Getenv("RATE_LIMIT_ENABLED"); rateLimitEnabled != "" {
		c.RateLimitEnabled = strings.ToLower(rateLimitEnabled) == "true"
	}

	// Ensure Auth config exists before setting values
	if c.Auth == nil {
		c.Auth = middleware.DefaultAuthConfig()
	}

	// AUTH_ENABLED
	if authEnabled := os.Getenv("AUTH_ENABLED"); authEnabled != "" {
		c.Auth.Enabled = strings.ToLower(authEnabled) == "true"
	}

	// AUTH_MODE
	if authMode := os.Getenv("AUTH_MODE"); authMode != "" {
		c.Auth.Mode = authMode
	}

	// BASIC_AUTH_USERNAME
	if username := os.Getenv("BASIC_AUTH_USERNAME"); username != "" {
		if c.Auth.BasicAuth == nil {
			c.Auth.BasicAuth = &middleware.BasicAuthConfig{}
		}
		c.Auth.BasicAuth.Username = username
	}

	// BASIC_AUTH_PASSWORD
	if password := os.Getenv("BASIC_AUTH_PASSWORD"); password != "" {
		if c.Auth.BasicAuth == nil {
			c.Auth.BasicAuth = &middleware.BasicAuthConfig{}
		}
		c.Auth.BasicAuth.Password = password
	}

	// TOKEN_AUTH_TOKENS (comma-separated list)
	if tokens := os.Getenv("TOKEN_AUTH_TOKENS"); tokens != "" {
		if c.Auth.TokenAuth == nil {
			c.Auth.TokenAuth = &middleware.TokenAuthConfig{}
		}
		c.Auth.TokenAuth.Tokens = strings.Split(tokens, ",")
	}
}

// Validate validates the configuration values
func (c *Config) Validate() error {
	log := utility.GetLogger()

	// Validate port range
	if c.Port < 1 || c.Port > 65535 {
		log.Error().Msgf("Invalid port number: %d", c.Port)
		return fmt.Errorf("configuration validation failed: port must be between 1 and 65535, got %d. Update 'port' in config file", c.Port)
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, strings.ToLower(c.LogLevel)) {
		log.Error().Msgf("Invalid log level: %s", c.LogLevel)
		return fmt.Errorf("configuration validation failed: log_level must be one of %v, got '%s'. Update 'log_level' in config file", validLogLevels, c.LogLevel)
	}

	// Validate worker pool size
	if c.WorkerPoolSize < 1 {
		log.Error().Msgf("Invalid worker pool size: %d", c.WorkerPoolSize)
		return fmt.Errorf("configuration validation failed: worker_pool_size must be at least 1, got %d. Increase 'worker_pool_size' in config file", c.WorkerPoolSize)
	}

	if c.WorkerPoolSize > 10000 {
		log.Error().Msgf("Worker pool size too large: %d", c.WorkerPoolSize)
		return fmt.Errorf("configuration validation failed: worker_pool_size must not exceed 10000, got %d (performance limit). Decrease 'worker_pool_size' in config file", c.WorkerPoolSize)
	}

	// Validate logs directory
	if c.LogsDir == "" {
		log.Error().Msg("logs_dir cannot be empty")
		return fmt.Errorf("configuration validation failed: logs_dir cannot be empty. Specify a directory path (e.g., './logs') in config file")
	}

	// Validate shutdown timeout
	if c.ShutdownTimeout < 1 {
		log.Error().Msgf("Invalid shutdown timeout: %d", c.ShutdownTimeout)
		return fmt.Errorf("configuration validation failed: shutdown_timeout must be at least 1 second, got %d. Increase 'shutdown_timeout' in config file", c.ShutdownTimeout)
	}
	if c.ShutdownTimeout > 300 {
		log.Warn().Msgf("Shutdown timeout is very high: %d seconds. Consider reducing it.", c.ShutdownTimeout)
	}

	// Validate rate limiting settings
	if c.RateLimitEnabled {
		if c.RateLimitRate <= 0 {
			log.Error().Msgf("Invalid rate limit rate: %f", c.RateLimitRate)
			return fmt.Errorf("configuration validation failed: rate_limit_rate must be greater than 0, got %f. Set a positive rate (requests per second) in config file", c.RateLimitRate)
		}
		if c.RateLimitBurst < 1 {
			log.Error().Msgf("Invalid rate limit burst: %d", c.RateLimitBurst)
			return fmt.Errorf("configuration validation failed: rate_limit_burst must be at least 1, got %d. Set burst capacity to handle traffic spikes in config file", c.RateLimitBurst)
		}
	}

	// Ensure Auth config exists
	if c.Auth == nil {
		c.Auth = middleware.DefaultAuthConfig()
	}

	// Validate authentication settings
	if c.Auth.Enabled {
		validAuthModes := []string{"basic", "token", "none"}
		if !contains(validAuthModes, strings.ToLower(c.Auth.Mode)) {
			log.Error().Msgf("Invalid auth mode: %s", c.Auth.Mode)
			return fmt.Errorf("configuration validation failed: auth.mode must be one of %v, got '%s'. Update 'auth.mode' in config file", validAuthModes, c.Auth.Mode)
		}

		// Validate basic auth configuration
		if strings.ToLower(c.Auth.Mode) == "basic" {
			if c.Auth.BasicAuth == nil || c.Auth.BasicAuth.Username == "" || c.Auth.BasicAuth.Password == "" {
				log.Error().Msg("Basic auth enabled but credentials not configured")
				return fmt.Errorf("configuration validation failed: auth.mode is 'basic' but username/password not provided. Configure 'auth.basic_auth.username' and 'auth.basic_auth.password' in config file")
			}
		}

		// Validate token auth configuration
		if strings.ToLower(c.Auth.Mode) == "token" {
			if c.Auth.TokenAuth == nil || len(c.Auth.TokenAuth.Tokens) == 0 {
				log.Error().Msg("Token auth enabled but no tokens configured")
				return fmt.Errorf("configuration validation failed: auth.mode is 'token' but no tokens provided. Configure 'auth.token_auth.tokens' in config file with at least one token")
			}
		}
	}

	return nil
}

// SaveToFile saves the current configuration to a YAML file
func (c *Config) SaveToFile(path string) error {
	log := utility.GetLogger()

	data, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Msgf("Failed to marshal config: %v", err)
		return fmt.Errorf("failed to serialize configuration to YAML format: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Error().Msgf("Failed to write config file '%s': %v", path, err)
		return fmt.Errorf("failed to write configuration to file '%s': %w. Check file permissions and disk space", path, err)
	}

	return nil
}

// GetServerAddr returns the server address in host:port format
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

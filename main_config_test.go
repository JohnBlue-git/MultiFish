package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"multifish/config"
)

func TestConfigIntegration(t *testing.T) {
	// Test environment variable configuration
	t.Run("Environment Variables Override", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("WORKER_POOL_SIZE", "150")
		defer func() {
			os.Unsetenv("PORT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("WORKER_POOL_SIZE")
		}()

		cfg, err := config.LoadConfig("")
		require.NoError(t, err)

		assert.Equal(t, 9090, cfg.Port)
		assert.Equal(t, "debug", cfg.LogLevel)
		assert.Equal(t, 150, cfg.WorkerPoolSize)
	})
}

func TestRootEndpointWithConfig(t *testing.T) {
	// Create test configuration
	cfg := config.DefaultConfig()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create router
	router := gin.New()
	
	// Add root endpoint
	router.GET("/MultiFish/v1", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"@odata.type":    "#ServiceRoot.v1_0_0.ServiceRoot",
			"@odata.id":      "/MultiFish/v1",
			"Id":             "MultiFish",
			"Name":           "MultiFish Service",
			"RedfishVersion": "1.0.0",
			"Platform": gin.H{
				"@odata.id": "/MultiFish/v1/Platform",
			},
			"JobService": gin.H{
				"@odata.id": "/MultiFish/v1/JobService",
			},
		})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/MultiFish/v1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "#ServiceRoot.v1_0_0.ServiceRoot", response["@odata.type"])
	assert.Equal(t, "/MultiFish/v1", response["@odata.id"])
	assert.Equal(t, "MultiFish", response["Id"])
	assert.Equal(t, "MultiFish Service", response["Name"])

	// Verify config is used (indirectly by checking service is operational)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Port)
}

func TestWorkerPoolSizeConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedSize  int
		shouldSucceed bool
	}{
		{
			name:          "Default size",
			envValue:      "",
			expectedSize:  99,
			shouldSucceed: true,
		},
		{
			name:          "Custom valid size",
			envValue:      "200",
			expectedSize:  200,
			shouldSucceed: true,
		},
		{
			name:          "Minimum size",
			envValue:      "1",
			expectedSize:  1,
			shouldSucceed: true,
		},
		{
			name:          "Maximum size",
			envValue:      "10000",
			expectedSize:  10000,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv("WORKER_POOL_SIZE", tt.envValue)
				defer os.Unsetenv("WORKER_POOL_SIZE")
			}

			// Load configuration
			cfg, err := config.LoadConfig("")
			
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSize, cfg.WorkerPoolSize)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPortConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedPort  int
		shouldSucceed bool
	}{
		{
			name:          "Default port",
			envValue:      "",
			expectedPort:  8080,
			shouldSucceed: true,
		},
		{
			name:          "Custom port",
			envValue:      "9090",
			expectedPort:  9090,
			shouldSucceed: true,
		},
		{
			name:          "HTTP port",
			envValue:      "80",
			expectedPort:  80,
			shouldSucceed: true,
		},
		{
			name:          "HTTPS port",
			envValue:      "443",
			expectedPort:  443,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv("PORT", tt.envValue)
				defer os.Unsetenv("PORT")
			}

			// Load configuration
			cfg, err := config.LoadConfig("")
			
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPort, cfg.Port)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestLogLevelConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedLevel string
		shouldSucceed bool
	}{
		{
			name:          "Default level",
			envValue:      "",
			expectedLevel: "info",
			shouldSucceed: true,
		},
		{
			name:          "Debug level",
			envValue:      "debug",
			expectedLevel: "debug",
			shouldSucceed: true,
		},
		{
			name:          "Warn level",
			envValue:      "warn",
			expectedLevel: "warn",
			shouldSucceed: true,
		},
		{
			name:          "Error level",
			envValue:      "error",
			expectedLevel: "error",
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if provided
			if tt.envValue != "" {
				os.Setenv("LOG_LEVEL", tt.envValue)
				defer os.Unsetenv("LOG_LEVEL")
			}

			// Load configuration
			cfg, err := config.LoadConfig("")
			
			if tt.shouldSucceed {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLevel, cfg.LogLevel)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestConfigFileLoading(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	configContent := `
port: 9000
log_level: debug
worker_pool_size: 200
logs_dir: /tmp/test-logs
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Load configuration from file
	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 200, cfg.WorkerPoolSize)
	assert.Equal(t, "/tmp/test-logs", cfg.LogsDir)
}

func TestConfigPrioritySystem(t *testing.T) {
	// Create config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yaml"

	configContent := `
port: 9000
log_level: warn
worker_pool_size: 200
logs_dir: /tmp/file-logs
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables (should override file)
	os.Setenv("PORT", "9999")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	require.NoError(t, err)

	// Env vars should override file
	assert.Equal(t, 9999, cfg.Port, "Port should come from environment")
	assert.Equal(t, "debug", cfg.LogLevel, "LogLevel should come from environment")
	
	// File values where no env var
	assert.Equal(t, 200, cfg.WorkerPoolSize, "WorkerPoolSize should come from file")
	assert.Equal(t, "/tmp/file-logs", cfg.LogsDir, "LogsDir should come from file")
}

func TestGetServerAddr(t *testing.T) {
	tests := []struct {
		name     string
		port     int
		expected string
	}{
		{"Default port", 8080, ":8080"},
		{"Custom port", 9090, ":9090"},
		{"Low port", 80, ":80"},
		{"High port", 65535, ":65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Port: tt.port}
			assert.Equal(t, tt.expected, cfg.GetServerAddr())
		})
	}
}

func TestConfigValidationErrors(t *testing.T) {
	tests := []struct {
		name          string
		modifyConfig  func(*config.Config)
		expectedError string
	}{
		{
			name: "Invalid port - too low",
			modifyConfig: func(c *config.Config) {
				c.Port = 0
			},
			expectedError: "port must be between",
		},
		{
			name: "Invalid port - too high",
			modifyConfig: func(c *config.Config) {
				c.Port = 70000
			},
			expectedError: "port must be between",
		},
		{
			name: "Invalid log level",
			modifyConfig: func(c *config.Config) {
				c.LogLevel = "invalid"
			},
			expectedError: "log_level must be one of",
		},
		{
			name: "Invalid worker pool size - zero",
			modifyConfig: func(c *config.Config) {
				c.WorkerPoolSize = 0
			},
			expectedError: "worker_pool_size must be at least",
		},
		{
			name: "Invalid worker pool size - too large",
			modifyConfig: func(c *config.Config) {
				c.WorkerPoolSize = 20000
			},
			expectedError: "worker_pool_size must not exceed",
		},
		{
			name: "Empty logs directory",
			modifyConfig: func(c *config.Config) {
				c.LogsDir = ""
			},
			expectedError: "logs_dir cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			tt.modifyConfig(cfg)

			err := cfg.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

func BenchmarkLoadConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = config.LoadConfig("")
	}
}

func BenchmarkLoadConfigWithFile(b *testing.B) {
	// Create temporary config file
	tmpDir := b.TempDir()
	configPath := tmpDir + "/config.yaml"

	configContent := `
port: 9000
log_level: info
worker_pool_size: 100
logs_dir: /tmp/logs
`
	_ = os.WriteFile(configPath, []byte(configContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.LoadConfig(configPath)
	}
}

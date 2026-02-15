package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 99, cfg.WorkerPoolSize)
	assert.Equal(t, "./logs", cfg.LogsDir)
	assert.Equal(t, 30, cfg.ShutdownTimeout)
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("WORKER_POOL_SIZE", "150")
	os.Setenv("LOGS_DIR", "/var/log/multifish")
	os.Setenv("SHUTDOWN_TIMEOUT", "60")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("WORKER_POOL_SIZE")
		os.Unsetenv("LOGS_DIR")
		os.Unsetenv("SHUTDOWN_TIMEOUT")
	}()

	cfg, err := LoadConfig("")
	require.NoError(t, err)

	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 150, cfg.WorkerPoolSize)
	assert.Equal(t, "/var/log/multifish", cfg.LogsDir)
	assert.Equal(t, 60, cfg.ShutdownTimeout)
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
port: 9000
log_level: warn
worker_pool_size: 200
logs_dir: /tmp/logs
shutdown_timeout: 45
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, 200, cfg.WorkerPoolSize)
	assert.Equal(t, "/tmp/logs", cfg.LogsDir)
	assert.Equal(t, 45, cfg.ShutdownTimeout)
}

func TestLoadConfigPriority(t *testing.T) {
	// Environment variables should override file values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
port: 9000
log_level: warn
worker_pool_size: 200
logs_dir: /tmp/logs
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set environment variables (should override file)
	os.Setenv("PORT", "9090")
	os.Setenv("WORKER_POOL_SIZE", "300")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("WORKER_POOL_SIZE")
	}()

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	// These should come from environment
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, 300, cfg.WorkerPoolSize)

	// These should come from file
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "/tmp/logs", cfg.LogsDir)
}

func TestLoadConfigFileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "configuration file not found")
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Invalid YAML
	invalidContent := `
port: not_a_number
log_level: [invalid
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	cfg, err := LoadConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to parse YAML")
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		expectErr bool
	}{
		{"Valid port 8080", 8080, false},
		{"Valid port 1", 1, false},
		{"Valid port 65535", 65535, false},
		{"Invalid port 0", 0, true},
		{"Invalid port negative", -1, true},
		{"Invalid port too high", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Port = tt.port

			err := cfg.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "port")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		expectErr bool
	}{
		{"Valid debug", "debug", false},
		{"Valid info", "info", false},
		{"Valid warn", "warn", false},
		{"Valid error", "error", false},
		{"Valid DEBUG (case insensitive)", "DEBUG", false},
		{"Invalid level", "invalid", true},
		{"Empty level", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LogLevel = tt.logLevel

			err := cfg.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "log_level")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateWorkerPoolSize(t *testing.T) {
	tests := []struct {
		name      string
		poolSize  int
		expectErr bool
	}{
		{"Valid size 1", 1, false},
		{"Valid size 99", 99, false},
		{"Valid size 10000", 10000, false},
		{"Invalid size 0", 0, true},
		{"Invalid size negative", -5, true},
		{"Invalid size too large", 10001, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.WorkerPoolSize = tt.poolSize

			err := cfg.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "worker_pool_size")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogsDir(t *testing.T) {
	tests := []struct {
		name      string
		logsDir   string
		expectErr bool
	}{
		{"Valid absolute path", "/var/log/multifish", false},
		{"Valid relative path", "./logs", false},
		{"Valid current dir", ".", false},
		{"Empty logs dir", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.LogsDir = tt.logsDir

			err := cfg.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "logs_dir")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateShutdownTimeout(t *testing.T) {
	tests := []struct {
		name            string
		shutdownTimeout int
		expectErr       bool
	}{
		{"Valid timeout 30s", 30, false},
		{"Valid timeout 60s", 60, false},
		{"Valid timeout 120s", 120, false},
		{"Minimum valid 1s", 1, false},
		{"High but valid 300s", 300, false},
		{"Zero timeout", 0, true},
		{"Negative timeout", -5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.ShutdownTimeout = tt.shutdownTimeout

			err := cfg.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "shutdown_timeout")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSaveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Port:            9000,
		LogLevel:        "debug",
		WorkerPoolSize:  150,
		LogsDir:         "/var/log/multifish",
		ShutdownTimeout: 45,
	}

	err := cfg.SaveToFile(configPath)
	require.NoError(t, err)

	// Load and verify
	loadedCfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Equal(t, cfg.Port, loadedCfg.Port)
	assert.Equal(t, cfg.LogLevel, loadedCfg.LogLevel)
	assert.Equal(t, cfg.WorkerPoolSize, loadedCfg.WorkerPoolSize)
	assert.Equal(t, cfg.LogsDir, loadedCfg.LogsDir)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Port: tt.port}
			assert.Equal(t, tt.expected, cfg.GetServerAddr())
		})
	}
}

func TestLoadConfigInvalidEnvValues(t *testing.T) {
	// Set invalid environment variables
	os.Setenv("PORT", "invalid")
	os.Setenv("WORKER_POOL_SIZE", "not_a_number")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("WORKER_POOL_SIZE")
	}()

	// Should still load with defaults (invalid env values are ignored)
	cfg, err := LoadConfig("")
	require.NoError(t, err)

	// Should use defaults since env values were invalid
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 99, cfg.WorkerPoolSize)
}

func TestLoadConfigMixedSources(t *testing.T) {
	// Create config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
port: 9000
log_level: warn
worker_pool_size: 200
logs_dir: /tmp/logs
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Set only some env variables
	os.Setenv("PORT", "9999")
	defer os.Unsetenv("PORT")

	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)

	// Port from env
	assert.Equal(t, 9999, cfg.Port)
	// Rest from file
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, 200, cfg.WorkerPoolSize)
	assert.Equal(t, "/tmp/logs", cfg.LogsDir)
}

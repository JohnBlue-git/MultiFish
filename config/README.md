# Config Package

Configuration management for MultiFish with support for environment variables and YAML files.

## Features

- **Environment Variables**: Runtime configuration via env vars
- **YAML Configuration**: File-based configuration for production
- **Priority System**: Env vars > YAML > Defaults
- **Validation**: Comprehensive validation of all settings
- **Type Safety**: Strongly-typed configuration structure

## Configuration Options

| Option | Env Variable | Default | Description |
|--------|--------------|---------|-------------|
| `Port` | `PORT` | `8080` | HTTP server port (1-65535) |
| `LogLevel` | `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `WorkerPoolSize` | `WORKER_POOL_SIZE` | `99` | Maximum concurrent jobs (1-10000) |
| `LogsDir` | `LOGS_DIR` | `./logs` | Directory for job execution logs |
| `ShutdownTimeout` | `SHUTDOWN_TIMEOUT` | `30` | Graceful shutdown timeout in seconds (1-300) |
| `RateLimitEnabled` | `RATE_LIMIT_ENABLED` | `true` | Enable/disable rate limiting |
| `RateLimitRate` | `RATE_LIMIT_RATE` | `10.0` | Requests per second per IP |
| `RateLimitBurst` | `RATE_LIMIT_BURST` | `20` | Maximum burst size |

## Usage

### Using Defaults

```go
import "multifish/config"

cfg, err := config.LoadConfig("")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Server will start on port %d\n", cfg.Port)
```

### Using Environment Variables

```bash
export PORT=9090
export LOG_LEVEL=debug
export WORKER_POOL_SIZE=150
export LOGS_DIR=/var/log/multifish
export SHUTDOWN_TIMEOUT=60

./multifish
```

Or inline:

```bash
PORT=9090 LOG_LEVEL=debug SHUTDOWN_TIMEOUT=60 ./multifish
```

### Using YAML Configuration File

Create `config.yaml`:

```yaml
port: 9090
log_level: debug
worker_pool_size: 150
logs_dir: /var/log/multifish
shutdown_timeout: 60
```

Load in application:

```go
cfg, err := config.LoadConfig("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### Using Command-Line Flag

```go
import (
    "flag"
    "multifish/config"
)

func main() {
    configPath := flag.String("config", "", "Path to configuration file")
    flag.Parse()

    cfg, err := config.LoadConfig(*configPath)
    if err != nil {
        log.Fatal(err)
    }

    // Use configuration
    addr := cfg.GetServerAddr()
}
```

Run with:

```bash
./multifish -config /etc/multifish/config.yaml
```

## Priority System

Configuration values are loaded in this order (later overrides earlier):

1. **Defaults** - Built-in defaults
2. **YAML File** - Values from configuration file
3. **Environment Variables** - Runtime environment settings

### Example

`config.yaml`:
```yaml
port: 9000
log_level: warn
worker_pool_size: 200
```

Environment:
```bash
export PORT=9090
export LOG_LEVEL=debug
```

Result:
- `Port`: `9090` (from env)
- `LogLevel`: `debug` (from env)
- `WorkerPoolSize`: `200` (from YAML)

## Validation Rules

### Port
- Must be between 1 and 65535
- Common ports: 80, 443, 8080, 9090

### LogLevel
- Must be one of: `debug`, `info`, `warn`, `error`
- Case-insensitive

### WorkerPoolSize
- Must be between 1 and 10000
- Recommended: 50-200 for production

### LogsDir
- Cannot be empty
- Can be absolute or relative path
- Directory is created automatically if it doesn't exist

### ShutdownTimeout
- Must be at least 1 second
- Maximum recommended: 300 seconds (5 minutes)
- **Development**: 30 seconds (default)
- **Production**: 60-120 seconds (for long-running jobs)
- Higher values allow more time for graceful cleanup during shutdown

### RateLimitRate
- Must be greater than 0
- Specifies requests per second per IP address
- Recommended: 5-10 for production, 10-20 for development

### RateLimitBurst
- Must be at least 1
- Maximum number of requests allowed in a burst
- Should be 2-3x the rate limit

## Production Configuration

### Example Production YAML

`/etc/multifish/config.yaml`:

```yaml
# HTTP Server Configuration
port: 8080

# Logging Configuration
log_level: info
logs_dir: /var/log/multifish

# Job Scheduler Configuration
worker_pool_size: 100

# Graceful Shutdown Configuration
shutdown_timeout: 60  # 60 seconds for production workloads

# Rate Limiting Configuration
rate_limit_enabled: true
rate_limit_rate: 5.0
rate_limit_burst: 10
```

### Docker Environment

`docker-compose.yml`:

```yaml
version: '3.8'
services:
  multifish:
    image: multifish:latest
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - WORKER_POOL_SIZE=100
      - LOGS_DIR=/app/logs
      - SHUTDOWN_TIMEOUT=60
      - RATE_LIMIT_ENABLED=true
      - RATE_LIMIT_RATE=5.0
      - RATE_LIMIT_BURST=10
    ports:
      - "8080:8080"
    volumes:
      - ./logs:/app/logs
```

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: multifish-config
data:
  config.yaml: |
    port: 8080
    log_level: info
    worker_pool_size: 100
    logs_dir: /var/log/multifish
    shutdown_timeout: 60
    rate_limit_enabled: true
    rate_limit_rate: 5.0
    rate_limit_burst: 10
---
apiVersion: v1
kind: Deployment
metadata:
  name: multifish
spec:
  template:
    spec:
      containers:
      - name: multifish
        image: multifish:latest
        env:
        - name: PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
        - name: SHUTDOWN_TIMEOUT
          value: "60"
        volumeMounts:
        - name: config
          mountPath: /etc/multifish
      volumes:
      - name: config
        configMap:
          name: multifish-config
```

## API Integration

The configuration is used throughout the application:

```go
// main.go
cfg, _ := config.LoadConfig(*configPath)

// Use port
srv := &http.Server{
    Addr:    cfg.GetServerAddr(),
    Handler: router,
}

// Use worker pool size
jobService = scheduler.NewJobService(cfg.WorkerPoolSize)

// Use logs directory
executor := scheduler.NewJobExecutor(cfg.LogsDir)

// Use shutdown timeout for graceful shutdown
shutdownTimeout := time.Duration(cfg.ShutdownTimeout) * time.Second
ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
defer cancel()
srv.Shutdown(ctx)
```

## Testing

Run configuration tests:

```bash
cd config
go test -v
go test -cover
```

Expected output:
```
=== RUN   TestDefaultConfig
--- PASS: TestDefaultConfig
=== RUN   TestLoadConfigFromEnv
--- PASS: TestLoadConfigFromEnv
...
PASS
coverage: 95.2% of statements
```

## Saving Configuration

Save current configuration to file:

```go
cfg := &config.Config{
    Port:            9090,
    LogLevel:        "debug",
    WorkerPoolSize:  150,
    LogsDir:         "/var/log/multifish",
    ShutdownTimeout: 60,
}

err := cfg.SaveToFile("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

## Configuration Best Practices

### Development Environment
```yaml
port: 8080
log_level: debug
worker_pool_size: 50
logs_dir: ./logs
shutdown_timeout: 30       # Shorter timeout for quick iterations
rate_limit_enabled: false  # Disable for easier testing
```

### Production Environment
```yaml
port: 8080
log_level: info
worker_pool_size: 100
logs_dir: /var/log/multifish
shutdown_timeout: 60       # Longer timeout for graceful job completion
rate_limit_enabled: true   # Enable for security
rate_limit_rate: 5.0
rate_limit_burst: 10
auth:
  enabled: true
  mode: token
```

### Shutdown Timeout Recommendations

| Scenario | Recommended Timeout | Reason |
|----------|-------------------|---------|
| Development | 30s | Fast iteration, minimal jobs |
| Testing | 30-45s | Allow test jobs to complete |
| Production (Light) | 60s | Standard graceful shutdown |
| Production (Heavy) | 90-120s | Long-running job operations |
| Batch Processing | 120-300s | Complex multi-step jobs |

**Note**: During shutdown, the server will:
1. Stop accepting new requests
2. Close machine connections
3. Stop the job scheduler
4. Wait up to `shutdown_timeout` seconds for in-flight requests to complete
5. Force shutdown if timeout is reached

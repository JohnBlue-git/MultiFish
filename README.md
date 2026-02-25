# MultiFish - Multi-BMC Redfish Management API

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**MultiFish** is a powerful REST API service for managing multiple Baseboard Management Controllers (BMCs) through the Redfish protocol. It provides centralized management, automated job scheduling, and support for both standard Redfish and vendor-specific extensions (OpenBMC).

## 🎯 Key Features

- **Multi-BMC Management**: Control multiple servers from a single API endpoint
- **Provider Architecture**: Support for standard Redfish and vendor-specific extensions
- **Job Scheduler**: Automate recurring tasks across multiple machines with flexible scheduling
- **Worker Pools**: Parallel execution with configurable concurrency control
- **Extensible Design**: Easy to add new BMC types, providers, and actions
- **Comprehensive Testing**: Unit and integration tests with coverage reporting
- **Structured Logging**: High-performance zerolog with JSON output and contextual fields
- **Detailed Execution Logs**: JSON execution logs for complete audit trails
- **Type-Safe Operations**: Strongly-typed payload validation and error handling
- **Security Features**: Flexible authentication (Basic/Token), rate limiting, and password masking

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Project Structure](#project-structure)
- [Core Components](#core-components)
- [Quick Start](#quick-start)
- [Security](#security)
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Job Scheduling](#job-scheduling)
- [Payload Examples](#payload-examples)
- [Configuration](#configuration)
- [Development](#development)
- [Testing](#testing)
- [Module Documentation](#module-documentation)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## 🏗️ Architecture Overview

MultiFish follows a modular, layered architecture designed for scalability and extensibility:

```
┌──────────────────────────────────────────────────────────────────┐
│                   REST API (Gin Framework)                       │
│      /MultiFish/v1/{Platform|JobService|Managers}               │
└────────────────┬─────────────────────────┬───────────────────────┘
                 │                         │
         ┌───────▼─────────┐       ┌──────▼────────────┐
         │  Platform Mgr   │       │  Job Scheduler    │
         │  (Connections)  │       │  (Automation)     │
         └───────┬─────────┘       └──────┬────────────┘
                 │                        │
                 │              ┌─────────▼────────────┐
                 │              │  Job Executor        │
                 │              │  (Worker Pools)      │
                 │              └─────────┬────────────┘
                 │                        │
         ┌───────▼────────────────────────▼─────────────┐
         │         Provider Registry                    │
         │  (Auto-detect BMC type and capabilities)     │
         └───────┬────────────────────────┬─────────────┘
                 │                        │
      ┌──────────▼────────┐     ┌─────────▼──────────┐
      │  Redfish Provider │     │  Extend Provider   │
      │  (Standard BMCs)  │     │  (OpenBMC + OEM)   │
      └──────────┬────────┘     └─────────┬──────────┘
                 │                        │
                 └──────────┬─────────────┘
                            │
                ┌───────────▼────────────┐
                │     BMC Hardware       │
                │  (Multiple Machines)   │
                └────────────────────────┘
```

**Architecture Highlights:**

- **API Layer**: RESTful endpoints following Redfish conventions
- **Platform Management**: Handles machine connections and discovery
- **Job Scheduling**: Time-based automation with worker pools
- **Provider System**: Pluggable architecture for different BMC types
- **Extensibility**: Easy to add new providers, actions, and features

## 📁 Project Structure

```
Gofish/
├── main.go                          # Application entry point
├── go.mod                           # Go module dependencies
├── go.sum                           # Dependency checksums
│
├── handler/                         # HTTP request handlers
│   ├── handlePlatform.go            # Platform/machine management
│   ├── handleJobService.go          # Job scheduling endpoints
│   ├── handleManager.go             # Manager operation endpoints
│   ├── handlePlatform_test.go       # Platform handler tests
│   └── handleJobService_test.go     # Job service handler tests
│
├── config/                          # Configuration management
│
├── config.example.yaml              # Example development config
├── config.production.yaml           # Example production config
├── .env.example                     # Environment variables example
│
├── providers/                       # Provider architecture
│
├── scheduler/                       # Job scheduling system
│
├── middleware/                      # HTTP middleware (auth, rate limiting)
│
├── utility/                         # Common utilities
│
├── tests/                           # Testing infrastructure
│
├── payloads/                        # Example payload files
│
├── examples.sh                      # Interactive API examples (Shell)
└── MultiFish.postman_collection.json # Postman collection (GUI)
```
## 🔧 Core Components

MultiFish is built around several core components that work together to provide comprehensive BMC management:

### 1. **Platform Management** ([Complete Documentation](handler/PLATFORM.md))

The foundation for all BMC interactions, handling connection lifecycle and machine registry.

**Key Features:**
- Multi-BMC connection management
- Service type abstraction (Base/Extend)
- Credential management and security
- HTTP connection pooling
- Automatic cleanup and resource management

**Service Types:**
- **Base Service** - Standard Redfish operations for cross-vendor compatibility
- **Extend Service** - OpenBMC with OEM extensions for advanced features

**What You Can Do:**
- Register and manage multiple BMC connections
- Configure timeouts and TLS settings
- Switch between service types dynamically
- Monitor connection health

📚 **[Read Full Platform Management Guide →](handler/PLATFORM.md)**

### 2. **Job Service** ([Complete Documentation](handler/JOBSERVICE.md))

Sophisticated scheduling system for automating BMC operations across multiple machines.

**Key Features:**
- Flexible scheduling (Once, Daily, Weekly, Monthly)
- Worker pool for concurrent execution (1-10000 workers)
- Comprehensive validation (schedule, payload, machines)
- Detailed JSON execution logs
- Automatic rescheduling for continuous jobs

**Supported Actions:**
- `PatchProfile` - Update thermal profiles
- `PatchManager` - Update manager properties
- `PatchFanController` - Configure fan controllers
- `PatchFanZone` - Manage fan zones
- `PatchPidController` - Tune PID controllers

**What You Can Do:**
- Schedule recurring BMC operations
- Execute actions across multiple machines
- Monitor job execution with detailed logs
- Dynamically adjust worker pool size

📚 **[Read Full Job Service Guide →](handler/JOBSERVICE.md)**

### 3. **Providers** ([Documentation](providers/README.md))

Pluggable architecture for different BMC types.

**Redfish Provider** - Standard Redfish operations:
- Manager metadata retrieval
- Basic property updates
- Cross-vendor compatibility

**Extend Provider** - OpenBMC with OEM extensions:
- Thermal profile management (Performance, Balanced, PowerSaver, Custom)
- Fan controller configuration
- Fan zone management
- PID controller tuning

**Provider Selection:**
```go
provider := managerProviders.FindProvider(manager)
// Automatically selects best provider based on capabilities
```

### 4. **Utility** ([Documentation](utility/README.md))

Common helper functions, error handling, and structured logging.

**Key Utilities:**
- **Structured Logging**: Zerolog-based logging with zero allocations
- **Log Levels**: trace, debug, info, warn, error, fatal, panic
- **Contextual Fields**: Add structured data to all log entries
- Redfish-compliant error responses
- Payload validation
- Type-safe conversions
- Unique ID generation

**Logging Features:**
- High-performance zero-allocation logging
- JSON-structured output for machine parsing
- Colored console output for development
- Automatic caller tracking (file:line)
- Configurable log levels per environment

### 5. **Configuration** ([Documentation](config/README.md))

Flexible configuration system with multiple sources.

**Configuration Sources:**
- Environment variables
- YAML configuration files
- Built-in defaults

**Configurable Options:**
- Server port
- Log level
- Worker pool size
- Logs directory

**Priority System:** Environment Variables > YAML File > Defaults

### 6. **Testing** ([Documentation](tests/README.md))

Comprehensive test infrastructure.

**Test Coverage:**
- Unit tests for all modules
- Integration tests for APIs
- Coverage reporting with HTML output
- Test result summaries

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or higher
- Access to one or more BMCs with Redfish support
- `jq` for pretty-printing JSON (optional)
- `curl` for API testing

### Installation

1. **Clone the repository:**
   ```bash
   cd /path/to/workspace
   git clone <repository-url>
   cd Gofish
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the application:**
   ```bash
   go build -o multifish
   ```

4. **Run the service:**

   **Option A: With defaults**
   ```bash
   ./multifish
   ```

   **Option B: With environment variables**
   ```bash
   PORT=9090 LOG_LEVEL=debug WORKER_POOL_SIZE=150 ./multifish
   ```

   **Option C: With configuration file**
   ```bash
   ./multifish -config config.yaml
   ```

   **Option D: With .env file**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   export $(cat .env | xargs)
   ./multifish
   ```

   The service starts on the configured port (default: `http://localhost:8080`).

### Verify Installation

```bash
curl http://localhost:8080/MultiFish/v1
```

Expected response:
```json
{
  "@odata.type": "#ServiceRoot.v1_0_0.ServiceRoot",
  "@odata.id": "/MultiFish/v1",
  "Id": "MultiFish",
  "Name": "MultiFish Service",
  "Platform": {
    "@odata.id": "/MultiFish/v1/Platform"
  },
  "JobService": {
    "@odata.id": "/MultiFish/v1/JobService"
  }
}
```

### Start Testing

**Option 1: Command Line (Shell Script)**
```bash
./examples.sh
# Follow the interactive menu
```

**Option 2: GUI (Postman)**
```
1. Install Postman (https://www.postman.com/downloads/)
2. Import: MultiFish.postman_collection.json
3. Click any request and hit "Send"
```

See [Usage Examples](#💡-usage-examples) for detailed instructions.

## � Security

MultiFish provides comprehensive security features to protect your API:

### Authentication & Authorization

Choose from multiple authentication modes:

- **No Authentication** (`none`) - For development/testing
- **Basic Authentication** (`basic`) - Username/password
- **Token Authentication** (`token`) - Bearer tokens (recommended for production)

**Quick Configuration:**

```yaml
# config.yaml
auth:
  enabled: true
  mode: token  # or "basic" or "none"
  token_auth:
    tokens:
      - "your-secret-token-here"
```

**Environment Variables:**

```bash
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="token1,token2,token3"
```

**Using the API with authentication:**

```bash
# Token authentication
curl -H "Authorization: Bearer your-token" \
     http://localhost:8080/MultiFish/v1

# Basic authentication
curl -u admin:password \
     http://localhost:8080/MultiFish/v1
```

### Rate Limiting

Protect against API abuse with configurable rate limiting:

```yaml
rate_limit_enabled: true
rate_limit_rate: 10.0    # Requests per second
rate_limit_burst: 20     # Burst capacity
```

### Additional Security Features

- **Password Masking**: BMC passwords are never exposed in API responses
- **Structured Logging**: Authentication failures are logged with IP addresses
- **HTTPS Support**: Use with reverse proxy (nginx, Caddy)

### Production Security Checklist

- [ ] Enable authentication (`auth.enabled: true`)
- [ ] Use token mode for production (`auth.mode: token`)
- [ ] Generate strong tokens: `openssl rand -hex 32`
- [ ] Enable rate limiting (`rate_limit_enabled: true`)
- [ ] Use HTTPS (reverse proxy)
- [ ] Store secrets in environment variables
- [ ] Monitor authentication failures in logs

**📚 Detailed Documentation:**

- **[Authentication Guide](docs/AUTHENTICATION.md)** - Complete authentication setup and examples
- **[SECURITY.md](SECURITY.md)** - Security features overview

> **Note:** Certificate-based authentication (mTLS) is not supported as we don't have a PKI infrastructure to verify client certificates. Use token authentication with a reverse proxy for certificate validation if needed.

## �📡 API Endpoints

### Platform Management

```
GET    /MultiFish/v1/Platform                    # List all Platform
POST   /MultiFish/v1/Platform                    # Register new platform
GET    /MultiFish/v1/Platform/{id}               # Get platform details
PATCH  /MultiFish/v1/Platform/{id}               # Update platform config
DELETE /MultiFish/v1/Platform/{id}               # Remove platform
GET    /MultiFish/v1/Platform/{id}/Systems       # List systems
GET    /MultiFish/v1/Platform/{id}/Managers      # List managers
```

**📚 See [PLATFORM.md](handler/PLATFORM.md) for detailed platform management documentation including:**
- Machine configuration options and validation
- Service types (Base vs Extend)
- Connection lifecycle management
- Security best practices
- Complete API reference with examples

### Manager Operations

```
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}           # Get manager
PATCH  /MultiFish/v1/Platform/{id}/Managers/{managerId}           # Update manager
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem       # Get OEM data
```

### OEM Extended Operations (OpenBMC)

```
# Profile Management
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/Profile
PATCH  /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/Profile

# Fan Controllers
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanControllers
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanControllers/{controllerId}
PATCH  /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanControllers/{controllerId}

# Fan Zones
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanZones
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanZones/{zoneId}
PATCH  /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/FanZones/{zoneId}

# PID Controllers
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/PidControllers
GET    /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/PidControllers/{controllerId}
PATCH  /MultiFish/v1/Platform/{id}/Managers/{managerId}/Oem/OpenBmc/Fan/PidControllers/{controllerId}
```

### Job Service

```
GET    /MultiFish/v1/JobService                   # Get service info
PATCH  /MultiFish/v1/JobService                   # Update configuration
GET    /MultiFish/v1/JobService/Jobs              # List all jobs
POST   /MultiFish/v1/JobService/Jobs              # Create new job
GET    /MultiFish/v1/JobService/Jobs/{jobId}      # Get job details
PATCH  /MultiFish/v1/JobService/Jobs/{jobId}      # Update job
DELETE /MultiFish/v1/JobService/Jobs/{jobId}      # Delete job
POST   /MultiFish/v1/JobService/Jobs/{jobId}/Actions/Trigger  # Trigger immediately
POST   /MultiFish/v1/JobService/Jobs/{jobId}/Actions/Cancel   # Cancel job
```

**📚 See [JOBSERVICE.md](handler/JOBSERVICE.md) for comprehensive job scheduling documentation including:**
- Schedule types (Once vs Continuous)
- Supported actions and payloads
- Worker pool configuration and sizing
- Execution flow and logging
- Complete examples and troubleshooting

## 💡 Usage Examples

You can interact with the MultiFish API in two ways:

### Method 1: Shell Script (Command Line)

Use the comprehensive examples script for automated testing and scripting:

```bash
# Interactive mode (menu-driven)
./examples.sh

# Direct execution
./examples.sh platform          # Platform management examples
./examples.sh manager           # Manager operations
./examples.sh profile           # Profile management
./examples.sh fan-controller    # Fan controller examples
./examples.sh job-create        # Job creation examples
./examples.sh all               # Run all examples
```

**Best for:**
- Automation and scripting
- CI/CD pipelines
- Quick command-line testing
- Shell script integration

### Method 2: Postman Collection (GUI)

Import the Postman collection for visual, interactive API testing:

#### **Setup Steps:**

1. **Install Postman**
   - Download from [postman.com](https://www.postman.com/downloads/)
   - Or use the web version

2. **Import Collection**
   ```
   1. Open Postman
   2. Click "Import" button
   3. Select "MultiFish.postman_collection.json"
   4. Collection appears in left sidebar
   ```

3. **Configure Variables**
   ```
   Collection Variables (click collection > Variables tab):
   - baseUrl: http://localhost:8080/MultiFish/v1
   - machineId: your-machine-id
   - managerId: bmc
   - fanControllerId: cpu_fan_controller
   ```

4. **Start Making Requests**
   - Click any request in the collection
   - Click "Send" button
   - View response in lower panel

#### **Available Requests:**

**Service Root**
- Get API information

**Platform Collection**
- List All Machines
- Add Machine
- Get Machine Details
- Update Machine Configuration
- Delete Machine

**Managers**
- Get Manager Details
- Update ServiceIdentification

**OEM OpenBMC**
- Profile: Get/Update (Acoustic, Performance, etc.)
- Fan Controllers: Get/Update

**Best for:**
- Visual API exploration
- Manual testing and debugging
- Team collaboration
- Learning the API structure
- Quick prototyping

#### **Example Workflow in Postman:**

```
1. Add a machine:
   Platform > Add Machine > Send

2. Get manager details:
   Managers > Get Manager Details > Send

3. Update profile:
   OEM OpenBMC > Profile > Update Profile - Performance > Send

4. Update fan controller:
   OEM OpenBMC > Fan Controllers > Update Fan Controller > Send
```

**💡 Tip:** You can also generate code snippets from Postman (Code button) for curl, Python, JavaScript, etc.

---

### API Examples (curl)

### 1. Register a Platform

```bash
curl -X POST http://localhost:8080/MultiFish/v1/Platform \
  -H "Content-Type: application/json" \
  -d '{
    "Id": "server-1",
    "Name": "Production Server 1",
    "Type": "Extend",
    "Endpoint": "https://192.168.1.100",
    "Username": "root",
    "Password": "password",
    "Insecure": true
  }'
```

### 2. Update Thermal Profile

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc/Oem/OpenBmc/Fan/Profile \
  -H "Content-Type: application/json" \
  -d '{
    "Profile": "Performance"
  }'
```

**Valid Profiles:**
- `Performance` - Maximum performance, higher power consumption
- `Balanced` - Optimal balance of performance and efficiency
- `PowerSaver` - Minimize power consumption
- `Custom` - User-defined settings

### 3. Configure Fan Controller

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc/Oem/OpenBmc/Fan/FanControllers/cpu_fan \
  -H "Content-Type: application/json" \
  -d '{
    "Multiplier": 1.2,
    "StepDown": 2,
    "StepUp": 5
  }'
```

### 4. Update Manager Properties

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1/Managers/bmc \
  -H "Content-Type: application/json" \
  -d '{
    "ServiceIdentification": "Production BMC v2.0"
  }'
```

## ⏰ Job Scheduling

The Job Service provides powerful automation for recurring BMC operations. Create jobs that execute on schedule across multiple machines.

### Quick Start

**Create a one-time job:**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d @payloads/patch_profile.json
```

**Create a daily recurring job:**
```bash
curl -X POST http://localhost:8080/MultiFish/v1/JobService/Jobs \
  -H "Content-Type: application/json" \
  -d @payloads/continuous_daily.json
```

### Common Use Cases

**Daily Power Management:**
- Switch to PowerSaver mode at night (10 PM)
- Switch to Performance mode in morning (8 AM)

**Weekday Workload:**
- Performance mode Monday-Friday at 8 AM
- Balanced mode for weekends

**Monthly Maintenance:**
- Update BMC configurations on 1st of month
- Reset counters on 15th of month

### Job Features

- ⏱️ **Flexible Scheduling**: Once, daily, weekly, monthly patterns
- 🔄 **Auto-Rescheduling**: Continuous jobs reschedule automatically
- 🚀 **Immediate Trigger**: Override schedule and run now
- 🔢 **Worker Pools**: Configurable concurrency (1-10000 workers)
- 📊 **Execution Logs**: Detailed JSON logs per execution
- ✅ **Validation**: Schedule, payload, and machine validation
- 🎯 **Multi-machine**: Execute across multiple BMCs simultaneously

**📚 For complete job scheduling documentation, see [JOBSERVICE.md](handler/JOBSERVICE.md)**

**Topics covered:**
- Schedule types and patterns (Once, Continuous, Daily, Weekly, Monthly)
- All supported actions (PatchProfile, PatchManager, PatchFanController, etc.)
- Payload structures and validation
- Worker pool sizing and configuration
- Execution flow and lifecycle
- Detailed troubleshooting guide
- Best practices and examples

## 📝 Payload Examples

All payload examples are available in the `payloads/` directory:

### Profile Update (`payloads/patch_profile.json`)

```json
{
  "Name": "Update Performance Profile",
  "Machines": ["machine-1", "machine-2"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {
        "Profile": "Performance"
      }
    }
  ],
  "Schedule": {
    "Type": "Once",
    "Time": "08:00:00",
    "Period": null
  }
}
```

### Multiple Managers (`payloads/patch_profile_multiple_managers.json`)

```json
{
  "Name": "Update Multiple Managers",
  "Machines": ["machine-1"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "Performance"}
    },
    {
      "ManagerID": "bmc2",
      "Payload": {"Profile": "Balanced"}
    }
  ],
  "Schedule": {
    "Type": "Once",
    "Time": "08:00:00",
    "Period": null
  }
}
```

### Manager Update (`payloads/patch_manager.json`)

```json
{
  "Name": "Update Manager Service Identification",
  "Machines": ["machine-1"],
  "Action": "PatchManager",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {
        "ServiceIdentification": "Production BMC v2.1"
      }
    }
  ],
  "Schedule": {
    "Type": "Once",
    "Time": "09:00:00",
    "Period": null
  }
}
```

### Fan Controller (`payloads/patch_fan_controller.json`)

```json
{
  "Name": "Update Fan Controller Settings",
  "Machines": ["machine-1"],
  "Action": "PatchFanController",
  "Payload": [
    {
      "ManagerID": "bmc",
      "FanControllerID": "cpu_fan_controller",
      "Payload": {
        "Multiplier": 1.2,
        "StepDown": 2,
        "StepUp": 5
      }
    }
  ],
  "Schedule": {
    "Type": "Once",
    "Time": "10:00:00",
    "Period": null
  }
}
```

### Daily Recurring (`payloads/continuous_daily.json`)

```json
{
  "Name": "Daily Profile Update",
  "Machines": ["machine-1"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "PowerSaver"}
    }
  ],
  "Schedule": {
    "Type": "Continuous",
    "Time": "22:00:00",
    "Period": {
      "StartDay": "2026-02-10",
      "EndDay": "2026-12-31",
      "DaysOfWeek": [],
      "DaysOfMonth": null
    }
  }
}
```

### Weekday Recurring (`payloads/continuous_weekdays.json`)

```json
{
  "Name": "Weekday Performance Mode",
  "Machines": ["machine-1"],
  "Action": "PatchProfile",
  "Payload": [
    {
      "ManagerID": "bmc",
      "Payload": {"Profile": "Performance"}
    }
  ],
  "Schedule": {
    "Type": "Continuous",
    "Time": "08:00:00",
    "Period": {
      "StartDay": "2026-02-10",
      "EndDay": "2026-12-31",
      "DaysOfWeek": ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"],
      "DaysOfMonth": null
    }
  }
}
```

See the [`payloads/`](payloads/) directory for all examples.

## ⚙️ Configuration

MultiFish supports multiple configuration sources with a clear priority system. See the [Config Documentation](config/README.md) for full details.

### Configuration Options

| Option | Env Variable | Default | Description |
|--------|--------------|---------|-------------|
| `Port` | `PORT` | `8080` | HTTP server port (1-65535) |
| `LogLevel` | `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |
| `WorkerPoolSize` | `WORKER_POOL_SIZE` | `99` | Maximum concurrent jobs (1-10000) |
| `LogsDir` | `LOGS_DIR` | `./logs` | Directory for job execution logs |

### Configuration Priority

Configuration is loaded in this order (later overrides earlier):

1. **Built-in Defaults** - Sensible defaults for all settings
2. **YAML Configuration File** - File-based configuration
3. **Environment Variables** - Runtime configuration (highest priority)

### Using Environment Variables

**Quick start:**
```bash
PORT=9090 LOG_LEVEL=debug WORKER_POOL_SIZE=150 ./multifish
```

**Using .env file:**
```bash
# Copy example and customize
cp .env.example .env

# Edit .env with your settings
nano .env

# Load and run
export $(cat .env | xargs)
./multifish
```

**Example .env:**
```bash
PORT=8080
LOG_LEVEL=info
WORKER_POOL_SIZE=99
LOGS_DIR=./logs
```

### Using YAML Configuration

**Development (config.example.yaml):**
```yaml
port: 8080
log_level: debug
worker_pool_size: 50
logs_dir: ./logs
```

**Production (config.production.yaml):**
```yaml
port: 8080
log_level: info
worker_pool_size: 100
logs_dir: /var/log/multifish
```

**Run with config file:**
```bash
./multifish -config config.yaml
```

### Docker Deployment

**docker-compose.yml:**
```yaml
version: '3.8'
services:
  multifish:
    build: .
    environment:
      - PORT=8080
      - LOG_LEVEL=info
      - WORKER_POOL_SIZE=100
      - LOGS_DIR=/app/logs
    ports:
      - "8080:8080"
    volumes:
      - ./logs:/app/logs
      - ./config.yaml:/app/config.yaml
    command: ./multifish -config config.yaml
```

### Kubernetes Deployment

**ConfigMap:**
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
```

**Deployment:**
```yaml
apiVersion: apps/v1
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
          valueFrom:
            configMapKeyRef:
              name: multifish-config
              key: log_level
        volumeMounts:
        - name: config
          mountPath: /etc/multifish
        - name: logs
          mountPath: /var/log/multifish
      volumes:
      - name: config
        configMap:
          name: multifish-config
      - name: logs
        emptyDir: {}
```

### Runtime Configuration Updates

Update worker pool size dynamically via API:

```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/JobService \
  -H "Content-Type: application/json" \
  -d '{
    "WorkerPoolSize": 150
  }'
```

**Note:** Only `WorkerPoolSize` can be updated at runtime. Other settings require restart.

### Configuration Best Practices

1. **Development**: Use defaults or `config.example.yaml`
2. **Production**: Use `config.production.yaml` or environment variables
3. **Containers**: Use environment variables for portability
4. **Kubernetes**: Use ConfigMaps and Secrets
5. **Security**: Never commit sensitive data to config files
6. **Validation**: Configuration is validated at startup

### Platform Configuration

When registering Platform:

```json
{
  "Id": "unique-id",
  "Name": "Display name",
  "Type": "Base|Extend",
  "Endpoint": "https://bmc-address",
  "Username": "admin",
  "Password": "password",
  "Insecure": true,
  "HTTPClientTimeout": 30,
  "DisableEtagMatch": true
}
```

**Platform Types:**
- **Base**: Standard Redfish BMC
- **Extend**: OpenBMC with OEM extensions

## � Logging

MultiFish uses [zerolog](https://github.com/rs/zerolog) for high-performance structured logging.

### Log Levels

Configure via `LOG_LEVEL` environment variable or `log_level` in config file:

- `trace` - Very detailed debugging (development only)
- `debug` - Debugging information
- `info` - General informational messages (recommended for production)
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (exits program)
- `panic` - Panic-level errors

### Log Output

**Console Output (Development):**
```
2:15PM INF Configuration loaded logLevel=debug port=8080 workerPoolSize=99
2:15PM INF MultiFish API server starting address=:8080
2:15PM INF Added machine machineID=machine1 endpoint=https://bmc1 type=ExtendService
2:15PM INF Job created jobID=Job-1707489234567890 nextRun=2024-02-09T20:00:00Z
2:15PM INF Executing job activeWorkers=1 jobID=Job-1707489234567890 poolSize=99
2:15PM INF Successfully executed action action=PatchProfile duration=1.2s jobID=Job-1 machineID=machine1
```

**JSON Output (Production):**
```json
{"level":"info","time":"2024-02-09T14:15:00Z","caller":"main.go:42","message":"Configuration loaded","port":8080,"logLevel":"info","workerPoolSize":99}
{"level":"info","time":"2024-02-09T14:15:01Z","caller":"main.go:88","message":"MultiFish API server starting","address":":8080"}
{"level":"info","time":"2024-02-09T14:16:00Z","caller":"handler/handlePlatform.go:140","message":"Added machine","machineID":"machine1","endpoint":"https://bmc1","type":"ExtendService"}
{"level":"info","time":"2024-02-09T14:17:00Z","caller":"job_service.go:118","message":"Job created","jobID":"Job-1707489234567890","nextRun":"2024-02-09T20:00:00Z"}
```

### Logging Best Practices

**Good - Structured with context:**
```go
log := utility.GetLogger()
log.Info().
    Str("jobID", job.ID).
    Int("machineCount", len(machines)).
    Str("action", string(action)).
    Msg("Job created")
```

**Avoid - Unstructured:**
```go
log.Printf("Job %s created with %d machines", job.ID, len(machines))
```

### Log Files

**Application Logs:**
- Console output (stdout/stderr)
- Configurable via log level

**Job Execution Logs:**
- Location: `logs/`
- Format: JSON files per execution
- Contains: Job details, machine results, errors, timing

**Example job log:**
```json
{
  "job_id": "job1",
  "machine_id": "machine1",
  "action": "PatchProfile",
  "timestamp": "2024-02-09T20:00:00Z",
  "status": "Success",
  "duration": "1.5s",
  "payload": {...}
}
```

### Production Logging Setup

**Recommended configuration:**
```yaml
# config.production.yaml
log_level: "info"  # or "warn" for reduced verbosity
```

**Capture logs to file:**
```bash
# Redirect to file
./multifish 2>&1 | tee multifish.log

# With log rotation (using logrotate)
./multifish >> /var/log/multifish/app.log 2>&1
```

**Docker logging:**
```yaml
services:
  multifish:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

**Centralized logging (e.g., ELK Stack):**
```bash
# Parse JSON logs
./multifish 2>&1 | filebeat -c filebeat.yml
```

### Debugging

Enable debug logging temporarily:
```bash
LOG_LEVEL=debug ./multifish
```

Or via API (requires restart):
```bash
# Edit config file and restart
sed -i 's/log_level: info/log_level: debug/' config.yaml
```

## �🔨 Development

### Project Setup

```bash
# Clone repository
git clone <repository-url>
cd Gofish

# Install dependencies
go mod download

# Run in development mode
go run main.go
```

### Code Structure

- **Handlers** (`handle*.go`): API endpoint implementations
- **Providers** (`providers/`): BMC type abstraction layer
- **Scheduler** (`scheduler/`): Job scheduling and execution
- **Utility** (`utility/`): Shared helper functions
- **Tests** (`tests/`, `*_test.go`): Testing infrastructure

### Adding a New Action

1. **Define action type** in `scheduler/job_action.go`:
   ```go
   const ActionNewAction ActionType = "NewAction"
   ```

2. **Create payload structure** in `scheduler/payload_models.go`:
   ```go
   type ExecuteNewActionPayload struct {
       ManagerID string
       Payload   NewActionType
   }
   ```

3. **Implement validation**:
   ```go
   func ValidateNewActionPayloads(payloads Payload) error {
       // Validation logic
   }
   ```

4. **Add execution handler** in `scheduler/job_action.go`:
   ```go
   func (dae *DefaultActionExecutor) ExecuteNewAction(machine interface{}, payload Payload) error {
       // Execution logic
   }
   ```

5. **Create example payload** in `payloads/new_action.json`

6. **Write tests** in `scheduler/*_test.go`

7. **Update documentation**

## 🧪 Testing

### Run All Tests

```bash
./tests/run_all_tests.sh
```

### Generate Coverage Report

```bash
./tests/coverage_report.sh
```

View HTML report:
```bash
open tests/reports/coverage_*.html
```

### Run Specific Tests

```bash
# Test a specific package
./tests/run_specific_test.sh multifish/scheduler

# Test a specific function
./tests/run_specific_test.sh multifish/scheduler TestJobCreation

# Test matching pattern
./tests/run_specific_test.sh multifish/scheduler TestJob.*
```

### Test Coverage Goals

| Module | Target | Status |
|--------|--------|--------|
| Config | 90%+ | ✅ |
| Scheduler | 80%+ | ✅ |
| Utility | 85%+ | ✅ |
| Providers | 75%+ | ✅ |
| Handlers | 70%+ | ✅ |

## 📚 Module Documentation

Comprehensive documentation for each module:

### Core Features
- **[Platform Management](handler/PLATFORM.md)** - BMC connection management, machine registration, service types, and API reference
- **[Job Service](handler/JOBSERVICE.md)** - Job scheduling, automation, worker pools, and execution logging

### Internal Modules
- **[Config](config/README.md)** - Configuration management and environment variables
- **[Providers](providers/README.md)** - Provider architecture and BMC type support
- **[Scheduler](scheduler/README.md)** - Job scheduling internals and implementation details
- **[Utility](utility/README.md)** - Helper functions, logging, and error handling
- **[Tests](tests/README.md)** - Testing infrastructure and guidelines

### Integration Guides
- **[Authentication](docs/AUTHENTICATION.md)** - Complete authentication setup and examples
- **[Security](SECURITY.md)** - Security features and best practices

## 🔍 Troubleshooting

### Service Won't Start

**Check:**
```bash
# Verify port is available
lsof -i :8080

# Check Go version
go version  # Should be 1.22+

# Verify dependencies
go mod download
```

### Platform Registration Fails

**Common issues:**
- BMC endpoint unreachable
- Invalid credentials
- Network firewall blocking connection
- SSL certificate issues (use `"Insecure": true` for testing)

**Debug:**
```bash
# Test BMC connectivity
curl -k https://bmc-address/redfish/v1

# Check MultiFish logs
# (Add logging to main.go if needed)
```

### Job Not Executing

**Check:**
1. Job status:
   ```bash
   curl http://localhost:8080/MultiFish/v1/JobService/Jobs/{jobId}
   ```

2. Verify NextRunTime is in the future
3. Check worker pool capacity:
   ```bash
   curl http://localhost:8080/MultiFish/v1/JobService
   ```

4. Review job logs in `logs/`

### Invalid Payload Error

**Common causes:**
- Unknown field in payload
- Invalid profile value
- Empty required fields
- Wrong payload type for action

**Solution:**
- Check payload examples in `payloads/` directory
- Review module documentation for allowed fields
- Validate JSON syntax

### Performance Issues

**Symptoms:**
- Slow API responses
- Jobs queuing up
- High CPU usage

**Solutions:**
1. Increase worker pool size
2. Reduce job frequency
3. Check BMC response times
4. Monitor system resources

## 🤝 Contributing

### Guidelines

1. **Code Style**: Follow Go conventions and existing patterns
2. **Testing**: Add tests for all new features
3. **Documentation**: Update relevant README files
4. **Commits**: Use clear, descriptive commit messages

### Development Workflow

1. Create feature branch
2. Implement changes with tests
3. Run test suite: `./tests/run_all_tests.sh`
4. Generate coverage: `./tests/coverage_report.sh`
5. Update documentation
6. Submit pull request

### Adding New Providers

See [Providers README](providers/README.md) for detailed guide on implementing new BMC type providers.

## 📄 License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Gofish](https://github.com/stmcginnis/gofish) - Redfish and Swordfish client library
- Uses [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- Inspired by Redfish specification from DMTF

## 📞 Support

For issues, questions, or contributions:
- Check the [module documentation](#module-documentation)
- Review [examples](examples.sh) and [payloads](payloads/)
- Read feature guides: [Platform Management](handler/PLATFORM.md) | [Job Service](handler/JOBSERVICE.md)
- Open an issue on the repository

## 🔗 Quick Links

### Feature Documentation
- **[Platform Management](handler/PLATFORM.md)** - Complete guide to managing BMC connections
  - Machine configuration and validation
  - Service types (Base vs Extend)
  - Connection lifecycle
  - API reference with examples
  
- **[Job Service](handler/JOBSERVICE.md)** - Comprehensive job scheduling guide
  - Schedule types and patterns
  - All supported actions
  - Worker pool configuration
  - Execution logs and troubleshooting

### Module Documentation
- [Config Module](config/README.md) - Configuration system
- [Providers Module](providers/README.md) - Provider architecture
- [Scheduler Module](scheduler/README.md) - Scheduling internals
- [Utility Module](utility/README.md) - Common utilities
- [Testing Guide](tests/README.md) - Test infrastructure

### Guides & Examples
- [Authentication Guide](docs/AUTHENTICATION.md) - Auth setup
- [Security Guide](SECURITY.md) - Security features
- [Examples Script](examples.sh) - Interactive CLI examples
- [Postman Collection](MultiFish.postman_collection.json) - GUI testing
- [Payload Examples](payloads/) - Job payload templates

---

**MultiFish** - Centralized BMC Management Made Simple

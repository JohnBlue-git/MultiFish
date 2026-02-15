# Platform Management

## Overview

The **Platform Management** module provides the core functionality for managing connections to multiple BMC (Baseboard Management Controller) devices through the Redfish protocol. It serves as the foundation for all BMC interactions in MultiFish, handling connection lifecycle, credential management, and service type abstraction.

## Table of Contents

- [Architecture](#architecture)
- [Core Concepts](#core-concepts)
- [Machine Configuration](#machine-configuration)
- [Service Types](#service-types)
- [Platform Manager](#platform-manager)
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Platform API Layer                         │
│   GET/POST/PATCH/DELETE /MultiFish/v1/Platform         │
└────────────────────┬────────────────────────────────────┘
                     │
         ┌───────────▼────────────┐
         │   PlatformManager      │
         │  - Connection Pool     │
         │  - Machine Registry    │
         │  - Lifecycle Mgmt      │
         └───────────┬────────────┘
                     │
     ┌───────────────┴────────────────┐
     │                                │
┌────▼────────┐              ┌───────▼────────┐
│ Base Service│              │ Extend Service │
│ (Standard)  │              │ (OpenBMC+OEM)  │
└────┬────────┘              └───────┬────────┘
     │                               │
     └──────────┬────────────────────┘
                │
        ┌───────▼────────┐
        │  Gofish Client │
        │ (HTTP/Redfish) │
        └───────┬────────┘
                │
        ┌───────▼────────┐
        │  BMC Hardware  │
        └────────────────┘
```

## Core Concepts

### Machine Connection

A **MachineConnection** represents an active connection to a BMC device and contains:

- **Configuration**: Connection parameters (endpoint, credentials, timeouts)
- **Client**: Gofish API client for HTTP/Redfish communication
- **Service**: Either BaseService or ExtendService based on machine type

```go
type MachineConnection struct {
    Config         MachineConfig
    Client         *gofish.APIClient
    BaseService    *gofish.Service        // For standard Redfish
    ExtendService  *extendprovider.ExtendService  // For OpenBMC + OEM
}
```

### Platform Manager

The **PlatformManager** is a singleton that manages all machine connections with:

- **Thread-safe operations**: RWMutex for concurrent access
- **Connection pooling**: Efficient reuse of HTTP connections
- **Lifecycle management**: Automatic cleanup and resource management
- **Registry**: Quick lookup by machine ID

```go
type PlatformManager struct {
    machines map[string]*MachineConnection
    mu       sync.RWMutex
}
```

**Key Features:**
- Concurrent-safe operations with read/write locks
- Automatic connection pooling and timeout management
- Graceful cleanup with `CleanupAll()`
- Password masking in API responses

## Machine Configuration

### Configuration Structure

```go
type MachineConfig struct {
    ID                    string   // Unique identifier (required)
    Name                  string   // Display name (optional)
    Type                  string   // "Base" or "Extend" (default: "Extend")
    TypeAllowableValues   []string // Valid types ["Base", "Extend"]
    Endpoint              string   // BMC URL (required)
    Username              string   // Login username (required)
    Password              string   // Login password (required)
    Insecure              bool     // Skip TLS verification
    HTTPClientTimeout     int      // Request timeout in seconds (default: 30)
    DisableEtagMatch      bool     // Disable ETag validation (default: false)
}
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ID` | string | ✅ Yes | - | Unique machine identifier |
| `Name` | string | ❌ No | - | Human-readable display name |
| `Type` | string | ❌ No | `"Extend"` | Service type: `"Base"` or `"Extend"` |
| `TypeAllowableValues` | []string | ❌ No | `["Base","Extend"]` | Valid type values |
| `Endpoint` | string | ✅ Yes | - | BMC Redfish API URL |
| `Username` | string | ✅ Yes | - | Authentication username |
| `Password` | string | ✅ Yes | - | Authentication password |
| `Insecure` | bool | ❌ No | `false` | Skip TLS certificate verification |
| `HTTPClientTimeout` | int | ❌ No | `30` | HTTP request timeout (seconds) |
| `DisableEtagMatch` | bool | ❌ No | `false` | Disable ETag-based conditional updates |

### Configuration Validation

The platform performs comprehensive validation:

1. **Required Fields**: ID, Endpoint, Username, Password must be present
2. **Type Validation**: Type must be in TypeAllowableValues
3. **Timeout**: HTTPClientTimeout must be positive if specified
4. **Endpoint**: Must be a valid URL format
5. **Connection Test**: Validates connectivity during registration

## Service Types

### Base Service (Standard Redfish)

**When to use:**
- Standard DMTF Redfish BMCs
- Generic server management
- Cross-vendor compatibility

**Capabilities:**
- Manager metadata retrieval
- Basic property updates
- System information access
- Standard Redfish resources

**Example:**
```json
{
  "Id": "standard-server",
  "Type": "Base",
  "Endpoint": "https://standard-bmc.example.com",
  "Username": "admin",
  "Password": "password",
  "Insecure": true
}
```

### Extend Service (OpenBMC + OEM)

**When to use:**
- OpenBMC-based systems
- Advanced thermal management needed
- Custom fan control required
- Vendor-specific features

**Additional Capabilities:**
- Thermal profile management (Performance, Balanced, PowerSaver, Custom)
- Fan controller configuration
- Fan zone management
- PID controller tuning
- OpenBMC OEM extensions

**Example:**
```json
{
  "Id": "openbmc-server",
  "Type": "Extend",
  "Endpoint": "https://openbmc.example.com",
  "Username": "root",
  "Password": "0penBmc",
  "Insecure": true,
  "HTTPClientTimeout": 45
}
```

### Service Type Comparison

| Feature | Base Service | Extend Service |
|---------|--------------|----------------|
| Standard Redfish | ✅ Yes | ✅ Yes |
| Manager Operations | ✅ Yes | ✅ Yes |
| System Information | ✅ Yes | ✅ Yes |
| Thermal Profiles | ❌ No | ✅ Yes |
| Fan Controllers | ❌ No | ✅ Yes |
| Fan Zones | ❌ No | ✅ Yes |
| PID Controllers | ❌ No | ✅ Yes |
| OpenBMC OEM | ❌ No | ✅ Yes |

## Platform Manager

### Core Operations

#### Add Machine

Registers a new BMC and establishes connection.

```go
func (pm *PlatformManager) AddMachine(config MachineConfig) error
```

**Process:**
1. Validates configuration
2. Sets defaults (timeout, type, allowable values)
3. Creates custom HTTP client with timeout
4. Establishes Redfish connection
5. Creates appropriate service (Base or Extend)
6. Registers in machine registry

**Error Handling:**
- Configuration validation failures
- Network connectivity issues
- Authentication failures
- TLS verification problems

#### Get Machine

Retrieves a machine connection by ID.

```go
func (pm *PlatformManager) GetMachine(id string) (*MachineConnection, error)
```

**Returns:**
- Machine connection if found
- Error if machine doesn't exist

**Thread Safety:**
- Uses read lock for concurrent access

#### List Machines

Returns all registered machines with passwords masked.

```go
func (pm *PlatformManager) ListMachines() []MachineConfig
```

**Features:**
- Automatic password masking
- Thread-safe iteration
- Returns configuration snapshots

#### Remove Machine

Removes a machine and cleans up resources.

```go
func (pm *PlatformManager) RemoveMachine(id string) error
```

**Cleanup Process:**
1. Validates machine exists
2. Logs out from Redfish session
3. Closes idle HTTP connections
4. Removes from registry

#### Cleanup All

Gracefully shuts down all connections.

```go
func (pm *PlatformManager) CleanupAll()
```

**Used for:**
- Application shutdown
- Testing cleanup
- Resource management

### Connection Management

#### HTTP Client Configuration

Custom HTTP client with optimized settings:

```go
transport := &http.Transport{
    TLSHandshakeTimeout: 10 * time.Second,
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

**Benefits:**
- Connection pooling for better performance
- Configurable timeouts per machine
- Automatic connection reuse
- Graceful handling of TLS

#### Insecure Mode

For development or self-signed certificates:

```go
if config.Insecure {
    transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}
```

**Security Warning:** Only use in trusted networks or for testing!

## API Endpoints

### GET /MultiFish/v1/Platform

List all registered machines.

**Request:**
```bash
curl http://localhost:8080/MultiFish/v1/Platform
```

**Response:**
```json
{
  "@odata.type": "#MachineCollection.MachineCollection",
  "@odata.id": "/MultiFish/v1/Platform",
  "Name": "Platform Machine Collection",
  "Members": [
    {
      "@odata.id": "/MultiFish/v1/Platform/server-1"
    },
    {
      "@odata.id": "/MultiFish/v1/Platform/server-2"
    }
  ],
  "Members@odata.count": 2
}
```

### POST /MultiFish/v1/Platform

Register a new machine.

**Request:**
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
    "Insecure": true,
    "HTTPClientTimeout": 30
  }'
```

**Response (201 Created):**
```json
{
  "@odata.id": "/MultiFish/v1/Platform/server-1",
  "Id": "server-1",
  "Name": "Production Server 1"
}
```

**Error Response (400 Bad Request):**
```json
{
  "error": {
    "code": "PropertyMissing",
    "message": "Machine ID is required"
  }
}
```

### GET /MultiFish/v1/Platform/{machineId}

Get machine details and available managers.

**Request:**
```bash
curl http://localhost:8080/MultiFish/v1/Platform/server-1
```

**Response:**
```json
{
  "@odata.type": "#Machine.v1_0_0.Machine",
  "@odata.id": "/MultiFish/v1/Platform/server-1",
  "Id": "server-1",
  "Name": "Production Server 1",
  "Type": "Extend",
  "Type@Redfish.AllowableValues": ["Base", "Extend"],
  "Description": "BMC Machine Resource",
  "Connection": {
    "Endpoint": "https://192.168.1.100",
    "Username": "root",
    "Password": "******",
    "Insecure": true,
    "HTTPClientTimeout": 30,
    "DisableEtagMatch": false
  },
  "Managers": {
    "@odata.id": "/MultiFish/v1/Platform/server-1/Managers",
    "Members": [
      {
        "@odata.id": "/MultiFish/v1/Platform/server-1/Managers/bmc"
      }
    ],
    "Members@odata.count": 1
  }
}
```

### PATCH /MultiFish/v1/Platform/{machineId}

Update machine configuration.

**Request:**
```bash
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1 \
  -H "Content-Type: application/json" \
  -d '{
    "Username": "admin",
    "HTTPClientTimeout": 60,
    "DisableEtagMatch": true
  }'
```

**Allowed Fields:**
- `Endpoint`
- `Username`
- `Password`
- `HTTPClientTimeout`
- `DisableEtagMatch`
- `Type`

**Response:**
```json
{
  "@odata.id": "/MultiFish/v1/Platform/server-1",
  "Id": "server-1",
  "Message": "Configuration updated successfully"
}
```

**Note:** Changes to Endpoint, Username, Password, or Type may require reconnection.

### DELETE /MultiFish/v1/Platform/{machineId}

Remove a machine from the platform.

**Request:**
```bash
curl -X DELETE http://localhost:8080/MultiFish/v1/Platform/server-1
```

**Response:**
```json
{
  "Message": "Machine server-1 removed successfully"
}
```

## Usage Examples

### Basic Registration

```bash
# Register a standard Redfish BMC
curl -X POST http://localhost:8080/MultiFish/v1/Platform \
  -H "Content-Type: application/json" \
  -d '{
    "Id": "dell-server-1",
    "Name": "Dell PowerEdge R640",
    "Type": "Base",
    "Endpoint": "https://dell-bmc.example.com",
    "Username": "admin",
    "Password": "secret123",
    "Insecure": false,
    "HTTPClientTimeout": 30
  }'
```

### OpenBMC Registration

```bash
# Register an OpenBMC system with extended features
curl -X POST http://localhost:8080/MultiFish/v1/Platform \
  -H "Content-Type: application/json" \
  -d '{
    "Id": "openbmc-1",
    "Name": "OpenBMC Development Server",
    "Type": "Extend",
    "Endpoint": "https://192.168.100.10",
    "Username": "root",
    "Password": "0penBmc",
    "Insecure": true,
    "HTTPClientTimeout": 45,
    "DisableEtagMatch": false
  }'
```

### Updating Configuration

```bash
# Change timeout and credentials
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/openbmc-1 \
  -H "Content-Type: application/json" \
  -d '{
    "Username": "administrator",
    "Password": "newPassword123",
    "HTTPClientTimeout": 60
  }'
```

### Switching Service Type

```bash
# Change from Base to Extend to enable OEM features
curl -X PATCH http://localhost:8080/MultiFish/v1/Platform/server-1 \
  -H "Content-Type: application/json" \
  -d '{
    "Type": "Extend"
  }'
```

### Bulk Registration (Script)

```bash
#!/bin/bash
# register_servers.sh

SERVERS=(
  "server-1:192.168.1.100"
  "server-2:192.168.1.101"
  "server-3:192.168.1.102"
)

for server in "${SERVERS[@]}"; do
  IFS=':' read -r id ip <<< "$server"
  
  curl -X POST http://localhost:8080/MultiFish/v1/Platform \
    -H "Content-Type: application/json" \
    -d "{
      \"Id\": \"$id\",
      \"Name\": \"Production Server $id\",
      \"Type\": \"Extend\",
      \"Endpoint\": \"https://$ip\",
      \"Username\": \"root\",
      \"Password\": \"password\",
      \"Insecure\": true
    }"
  
  echo "Registered $id at $ip"
done
```

## Best Practices

### 1. Machine ID Naming

**Good:**
```
server-prod-01
datacenter-a-rack-5-node-3
dell-r640-bmc-10.0.1.50
```

**Avoid:**
```
1
temp
test-123-final-final-v2
```

**Recommendations:**
- Use descriptive, hierarchical names
- Include location or function information
- Keep IDs consistent across environments
- Avoid special characters (use hyphens, not underscores or spaces)

### 2. Timeout Configuration

```bash
# Production servers (stable network)
"HTTPClientTimeout": 30

# Remote servers (high latency)
"HTTPClientTimeout": 60

# Local development
"HTTPClientTimeout": 15
```

**Guidelines:**
- Start with default (30 seconds)
- Increase for slow/remote BMCs
- Decrease for local/fast networks
- Monitor timeout errors in logs

### 3. Security Considerations

**Development:**
```json
{
  "Insecure": true,
  "Password": "simple-password"
}
```

**Production:**
```json
{
  "Insecure": false,
  "Password": "Complex-P@ssw0rd-2024!"
}
```

**Best Practices:**
- Use strong passwords (min 12 chars, mixed case, numbers, symbols)
- Enable TLS verification in production (`Insecure: false`)
- Store passwords in environment variables or secrets manager
- Rotate credentials regularly
- Use separate accounts for MultiFish (not root)

### 4. ETag Management

```json
{
  "DisableEtagMatch": false  // Enable optimistic locking (recommended)
}
```

**When to disable:**
- BMC doesn't support ETags properly
- Experiencing ETag mismatch errors
- Need to force updates regardless of state

**When to enable (default):**
- Production systems
- Prevent concurrent update conflicts
- Ensure data consistency

### 5. Error Handling

```bash
# Check registration status
response=$(curl -s -w "\n%{http_code}" -X POST ...)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -ne 201 ]; then
  echo "Registration failed: $body"
  exit 1
fi
```

### 6. Connection Lifecycle

```go
// Application shutdown
defer platformMgr.CleanupAll()
```

**Important:**
- Always call CleanupAll() on shutdown
- Remove machines before decommissioning BMCs
- Monitor for connection leaks in long-running processes

## Troubleshooting

### Connection Failures

**Problem:** `failed to connect to <endpoint>`

**Solutions:**
1. **Check network connectivity:**
   ```bash
   ping <bmc-ip>
   curl -k https://<bmc-ip>/redfish/v1
   ```

2. **Verify credentials:**
   - Test login via BMC web interface
   - Check username/password spelling
   - Ensure account is not locked

3. **Check TLS settings:**
   - If self-signed cert: `"Insecure": true`
   - Verify BMC certificate is valid
   - Check client TLS version compatibility

4. **Increase timeout:**
   ```json
   "HTTPClientTimeout": 60
   ```

### Invalid Type Errors

**Problem:** `invalid Type 'XYZ', must be one of: [Base, Extend]`

**Solution:**
```json
{
  "Type": "Extend"  // Must be exactly "Base" or "Extend"
}
```

### Machine Not Found

**Problem:** `machine xyz not found`

**Diagnosis:**
```bash
# List all machines
curl http://localhost:8080/MultiFish/v1/Platform

# Check specific machine
curl http://localhost:8080/MultiFish/v1/Platform/xyz
```

**Common causes:**
- Typo in machine ID
- Machine was removed
- Wrong MultiFish instance

### Timeout Issues

**Problem:** Requests timing out

**Solutions:**
1. **Increase timeout:**
   ```bash
   curl -X PATCH .../Platform/server-1 \
     -d '{"HTTPClientTimeout": 90}'
   ```

2. **Check BMC performance:**
   - BMC may be overloaded
   - Network latency issues
   - BMC firmware problems

3. **Monitor logs:**
   ```bash
   # Enable debug logging
   LOG_LEVEL=debug ./multifish
   ```

### Password Exposure

**Problem:** Password visible in logs or responses

**Expected Behavior:**
- Passwords are automatically masked in GET responses: `"Password": "******"`
- Logs never contain passwords
- POST/PATCH requests may show passwords in debug logs

**If passwords appear:**
- Check utility.MaskPassword() is used
- Verify structured logging implementation
- Review debug log settings

### Resource Leaks

**Problem:** Too many open connections

**Solutions:**
1. **Call CleanupAll() on shutdown:**
   ```go
   defer platformMgr.CleanupAll()
   ```

2. **Remove unused machines:**
   ```bash
   curl -X DELETE .../Platform/old-machine
   ```

3. **Monitor connections:**
   ```bash
   netstat -an | grep :443 | wc -l
   ```

### Performance Issues

**Problem:** Slow machine registration or operations

**Solutions:**
1. **Enable HTTP connection pooling:** (already enabled by default)
2. **Reduce timeout for faster failures:**
   ```json
   "HTTPClientTimeout": 15
   ```

3. **Use concurrent operations in scripts:**
   ```bash
   # Register machines in parallel
   for server in "${SERVERS[@]}"; do
     register_machine "$server" &
   done
   wait
   ```

## Integration with Job Scheduler

The Platform Manager is tightly integrated with the Job Scheduler:

### PlatformManagerAdapter

```go
type PlatformManagerAdapter struct {
    mgr *PlatformManager
}

func (pma *PlatformManagerAdapter) GetMachine(machineID string) (interface{}, error) {
    return pma.mgr.GetMachineInterface(machineID)
}
```

**Purpose:**
- Provides scheduler.JobPlatformManager interface
- Enables job execution across machines
- Abstracts machine retrieval for validators and executors

### Usage in Jobs

```json
{
  "Name": "Update Multiple Servers",
  "Machines": ["server-1", "server-2", "server-3"],
  "Action": "PatchProfile",
  "Payload": [...],
  "Schedule": {...}
}
```

**Process:**
1. Job scheduler requests machines via adapter
2. Platform manager returns machine connections
3. Job executor performs actions on machines
4. Results logged per machine

See [JOBSERVICE.md](JOBSERVICE.md) for detailed job scheduling documentation.

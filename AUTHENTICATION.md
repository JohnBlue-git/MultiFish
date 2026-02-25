# Authentication & Authorization Configuration Guide

## Overview

MultiFish supports flexible authentication to secure your API endpoints. You can choose from three authentication modes:

1. **No Authentication (`none`)** - API is publicly accessible (default for development)
2. **Basic Authentication (`basic`)** - Traditional username/password authentication
3. **Token Authentication (`token`)** - Bearer token-based authentication

## Table of Contents

- [Quick Start](#quick-start)
- [Authentication Modes](#authentication-modes)
- [Configuration Methods](#configuration-methods)
- [Examples](#examples)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)
- [Limitations](#limitations)

---

## Quick Start

### Disable Authentication (Development)

```yaml
# config.yaml
auth:
  enabled: false
  mode: none
```

### Enable Basic Authentication

```yaml
# config.yaml
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "admin"
    password: "your-secure-password"
```

### Enable Token Authentication

```yaml
# config.yaml
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "my-secret-token-123"
      - "another-valid-token-456"
```

---

## Authentication Modes

### 1. No Authentication (`none`)

**When to use:**
- Local development
- Internal networks with other security measures
- Testing environments

**Configuration:**
```yaml
auth:
  enabled: false
  mode: none
```

**⚠️ Warning:** Not recommended for production environments. API will be publicly accessible.

---

### 2. Basic Authentication (`basic`)

**When to use:**
- Simple authentication needs
- Internal tools
- Quick prototyping with authentication

**How it works:**
- Client sends credentials in HTTP `Authorization` header
- Format: `Authorization: Basic base64(username:password)`
- Server validates against configured username/password

**Configuration:**

Via YAML:
```yaml
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "admin"
    password: "SecurePassword123!"
```

Via Environment Variables:
```bash
export AUTH_ENABLED=true
export AUTH_MODE=basic
export BASIC_AUTH_USERNAME=admin
export BASIC_AUTH_PASSWORD=SecurePassword123!
```

**Example API Request:**

Using curl:
```bash
curl -u admin:SecurePassword123! http://localhost:8080/MultiFish/v1
```

Using Authorization header:
```bash
curl -H "Authorization: Basic YWRtaW46U2VjdXJlUGFzc3dvcmQxMjMh" \
     http://localhost:8080/MultiFish/v1
```

Using Python:
```python
import requests
from requests.auth import HTTPBasicAuth

response = requests.get(
    'http://localhost:8080/MultiFish/v1',
    auth=HTTPBasicAuth('admin', 'SecurePassword123!')
)
```

**Pros:**
- Simple to implement
- Widely supported
- Built into most HTTP clients

**Cons:**
- Credentials sent with every request
- Single username/password (no multi-user support)
- Less flexible than token-based auth

---

### 3. Token Authentication (`token`)

**When to use:**
- Production environments
- Multiple clients/users
- Integration with other systems
- API key management

**How it works:**
- Client sends token in HTTP `Authorization` header
- Format: `Authorization: Bearer <token>`
- Server validates against list of configured tokens

**Configuration:**

Via YAML:
```yaml
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "production-token-abc123xyz"
      - "backup-token-def456uvw"
      - "integration-token-ghi789rst"
```

Via Environment Variables:
```bash
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="token1,token2,token3"
```

**Example API Request:**

Using curl:
```bash
curl -H "Authorization: Bearer production-token-abc123xyz" \
     http://localhost:8080/MultiFish/v1
```

Using Python:
```python
import requests

headers = {
    'Authorization': 'Bearer production-token-abc123xyz'
}

response = requests.get(
    'http://localhost:8080/MultiFish/v1',
    headers=headers
)
```

Using JavaScript (fetch):
```javascript
fetch('http://localhost:8080/MultiFish/v1', {
  headers: {
    'Authorization': 'Bearer production-token-abc123xyz'
  }
})
.then(response => response.json())
.then(data => console.log(data));
```

**Generating Secure Tokens:**

Linux/macOS:
```bash
# Generate a random 32-character token
openssl rand -hex 32

# Or using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

**Pros:**
- Multiple tokens for different clients/users
- Can revoke individual tokens
- No password transmission
- Better for automation and integrations

**Cons:**
- Tokens must be stored securely
- No built-in expiration (manual rotation required)

---

## Configuration Methods

### Priority Order

Configuration values are loaded in this priority (highest to lowest):

1. **Environment Variables** (highest priority)
2. **YAML Configuration File**
3. **Default Values** (lowest priority)

### Method 1: YAML Configuration File

Create or edit `config.yaml`:

```yaml
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "admin"
    password: "secret"
  token_auth:
    tokens: []
```

Run with config file:
```bash
./multifish -config config.yaml
```

### Method 2: Environment Variables

Set environment variables (overrides YAML):

```bash
# Enable/disable authentication
export AUTH_ENABLED=true

# Set authentication mode
export AUTH_MODE=basic

# Basic auth credentials
export BASIC_AUTH_USERNAME=admin
export BASIC_AUTH_PASSWORD=secret123

# Token auth (comma-separated list)
export TOKEN_AUTH_TOKENS=token1,token2,token3

# Run the application
./multifish
```

### Method 3: Systemd Service (Production)

Create `/etc/systemd/system/multifish.service`:

```ini
[Unit]
Description=MultiFish API Service
After=network.target

[Service]
Type=simple
User=multifish
WorkingDirectory=/opt/multifish
ExecStart=/opt/multifish/multifish -config /etc/multifish/config.yaml

# Authentication via environment variables
Environment="AUTH_ENABLED=true"
Environment="AUTH_MODE=token"
Environment="TOKEN_AUTH_TOKENS=your-production-token-here"

Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable multifish
sudo systemctl start multifish
```

### Method 4: Docker Environment Variables

```bash
docker run -d \
  -p 8080:8080 \
  -e AUTH_ENABLED=true \
  -e AUTH_MODE=token \
  -e TOKEN_AUTH_TOKENS=docker-token-123 \
  multifish:latest
```

---

## Examples

### Example 1: Development Without Authentication

```yaml
# config.development.yaml
port: 8080
log_level: debug
auth:
  enabled: false
  mode: none
```

```bash
./multifish -config config.development.yaml
curl http://localhost:8080/MultiFish/v1
```

### Example 2: Production with Token Auth

```yaml
# config.production.yaml
port: 8080
log_level: info
auth:
  enabled: true
  mode: token
  token_auth:
    tokens: []  # Set via environment variable for security
```

```bash
export TOKEN_AUTH_TOKENS="prod-token-$(openssl rand -hex 32)"
./multifish -config config.production.yaml
```

Test:
```bash
curl -H "Authorization: Bearer prod-token-..." \
     http://localhost:8080/MultiFish/v1
```

### Example 3: Basic Auth for Internal Tool

```yaml
# config.internal.yaml
port: 8080
log_level: info
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "ops-team"
    password: "internal-password-123"
```

### Example 4: Multiple Tokens for Different Services

```yaml
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "monitoring-service-token-abc"
      - "dashboard-token-def"
      - "automation-token-ghi"
      - "backup-token-jkl"
```

---

## Security Best Practices

### 🔒 General Security

1. **Enable Authentication in Production**
   ```yaml
   auth:
     enabled: true  # Always true in production
   ```

2. **Use Environment Variables for Secrets**
   - Never commit passwords/tokens to version control
   - Use environment variables or secret management systems
   ```bash
   export BASIC_AUTH_PASSWORD=$(cat /secure/password.txt)
   ```

3. **Use HTTPS in Production**
   - Use a reverse proxy (nginx, Apache, Caddy)
   - Terminate SSL/TLS at the proxy level
   ```nginx
   server {
       listen 443 ssl;
       ssl_certificate /path/to/cert.pem;
       ssl_certificate_key /path/to/key.pem;
       
       location / {
           proxy_pass http://localhost:8080;
       }
   }
   ```

### 🔑 Token Security

1. **Generate Strong Tokens**
   ```bash
   # At least 32 characters
   openssl rand -hex 32
   ```

2. **Rotate Tokens Regularly**
   - Add new token
   - Update clients
   - Remove old token

3. **Use Different Tokens for Different Clients**
   ```yaml
   token_auth:
     tokens:
       - "client-a-token"
       - "client-b-token"
       - "client-c-token"
   ```

4. **Store Tokens Securely**
   - Use secret management (Vault, AWS Secrets Manager)
   - Encrypt at rest
   - Never log tokens

### 🔐 Password Security

1. **Use Strong Passwords**
   - Minimum 12 characters
   - Mix of letters, numbers, symbols
   ```bash
   # Generate strong password
   openssl rand -base64 24
   ```

2. **Change Default Credentials**
   - Never use example passwords in production
   - Update immediately after deployment

### 📝 Logging

- Authentication failures are logged with IP address
- Successful authentications are not logged (to avoid token exposure)
- Monitor logs for suspicious activity:
  ```bash
  journalctl -u multifish | grep "Unauthorized"
  ```

---

## Troubleshooting

### Authentication Not Working

**Symptom:** Getting 401 Unauthorized errors

**Checklist:**
1. Verify authentication is enabled:
   ```bash
   curl http://localhost:8080/MultiFish/v1
   # Should return 401 if auth is enabled
   ```

2. Check configuration:
   ```bash
   # View logs
   journalctl -u multifish -n 50
   
   # Look for: "Authentication enabled" message
   ```

3. Verify credentials format:
   - Basic Auth: `Authorization: Basic base64(username:password)`
   - Token Auth: `Authorization: Bearer <token>`

4. Test with curl verbose mode:
   ```bash
   curl -v -H "Authorization: Bearer your-token" \
        http://localhost:8080/MultiFish/v1
   ```

### Environment Variables Not Applied

**Solution:**
- Restart the service after changing environment variables
- Check variable names (case-sensitive)
- Verify systemd service file if using systemd

### 401 Unauthorized Despite Correct Credentials

**Possible causes:**
1. Token has extra spaces or newlines
2. Wrong authentication mode configured
3. Token not in configured list
4. Environment variable override

**Debug:**
```bash
# Check exact token value
echo -n "your-token" | xxd

# Test without environment variables
unset AUTH_ENABLED AUTH_MODE TOKEN_AUTH_TOKENS
./multifish -config config.yaml
```

### Configuration Validation Errors

**Symptom:** Server fails to start with validation error

**Common errors:**

1. `auth.mode is 'basic' but username/password not provided`
   - Solution: Set username and password when using basic mode

2. `auth.mode is 'token' but no tokens provided`
   - Solution: Provide at least one token when using token mode

3. `auth.mode must be one of [basic token none]`
   - Solution: Use valid mode name

---

## Limitations

### What Authentication Does NOT Support

1. **Certificate-Based Authentication (mTLS)**
   - Not supported
   - Reason: No PKI infrastructure to verify client certificates
   - Alternative: Use token authentication with secure token management

2. **OAuth 2.0 / OpenID Connect**
   - Not supported
   - Use token authentication as alternative

3. **JWT (JSON Web Tokens)**
   - Not supported (opaque tokens only)
   - No token expiration/refresh mechanism
   - Manual token rotation required

4. **Multi-User Management**
   - Basic auth: Single username/password
   - Token auth: No user tracking (tokens are anonymous)

5. **Role-Based Access Control (RBAC)**
   - All authenticated users have full access
   - No per-endpoint or per-resource permissions

6. **Rate Limiting Per User**
   - Rate limiting is per-IP, not per-user/token

### Why Certificate Authentication is Not Supported

**Certificate authentication (mTLS) requires:**
- Certificate Authority (CA) infrastructure
- Certificate issuance and revocation system
- Certificate validation service
- Regular certificate renewal process

**MultiFish is designed as:**
- Lightweight API service
- Simple deployment model
- No external dependencies for auth

**For certificate-based security:**
- Use a reverse proxy with mTLS support (nginx, Envoy)
- Proxy validates certificates
- Forwards authenticated requests to MultiFish
- Example:
  ```nginx
  server {
      listen 443 ssl;
      ssl_client_certificate /path/to/ca.crt;
      ssl_verify_client on;
      
      location / {
          proxy_pass http://localhost:8080;
      }
  }
  ```

---

## API Response Examples

### Successful Request (Authenticated)

```bash
curl -H "Authorization: Bearer valid-token" \
     http://localhost:8080/MultiFish/v1
```

Response (200 OK):
```json
{
  "@odata.type": "#ServiceRoot.v1_0_0.ServiceRoot",
  "@odata.id": "/MultiFish/v1",
  "Id": "MultiFish",
  "Name": "MultiFish Service",
  "RedfishVersion": "1.0.0",
  "Platform": {
    "@odata.id": "/MultiFish/v1/Platform"
  },
  "JobService": {
    "@odata.id": "/MultiFish/v1/JobService"
  }
}
```

### Failed Request (No Credentials)

```bash
curl http://localhost:8080/MultiFish/v1
```

Response (401 Unauthorized):
```json
{
  "error": "Unauthorized",
  "message": "Valid Bearer token required"
}
```

### Failed Request (Invalid Credentials)

```bash
curl -H "Authorization: Bearer invalid-token" \
     http://localhost:8080/MultiFish/v1
```

Response (401 Unauthorized):
```json
{
  "error": "Unauthorized",
  "message": "Valid Bearer token required"
}
```

### Failed Request (Basic Auth - No Credentials)

Response includes WWW-Authenticate header:
```
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Basic realm="MultiFish API"
```

```json
{
  "error": "Unauthorized",
  "message": "Valid Basic Authentication credentials required"
}
```

---

## Migration Guide

### Upgrading from No Auth to Token Auth

1. **Generate tokens:**
   ```bash
   TOKEN1=$(openssl rand -hex 32)
   TOKEN2=$(openssl rand -hex 32)
   ```

2. **Update configuration:**
   ```yaml
   auth:
     enabled: true
     mode: token
     token_auth:
       tokens:
         - "${TOKEN1}"
         - "${TOKEN2}"
   ```

3. **Distribute tokens to clients**

4. **Deploy and test:**
   ```bash
   # Test with new token
   curl -H "Authorization: Bearer ${TOKEN1}" \
        http://localhost:8080/MultiFish/v1
   ```

5. **Monitor logs for auth failures**

---

## Related Documentation

- [Main README](README.md) - General information and setup
- [SECURITY.md](SECURITY.md) - Overall security features
- [Configuration Guide](config/README.md) - Full configuration reference

---

## Support

For issues or questions:
1. Check logs: `journalctl -u multifish -f`
2. Verify configuration validation passes
3. Test with curl verbose mode: `curl -v ...`
4. Review this guide's troubleshooting section

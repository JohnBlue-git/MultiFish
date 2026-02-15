# Security Improvements Documentation

This document describes the security enhancements implemented in the MultiFish application.

## Overview

Three major security improvements have been implemented:
1. **Authentication & Authorization**: Flexible authentication with multiple modes
2. **Password Masking**: Ensures passwords are never exposed in API responses
3. **Rate Limiting**: Protects the API from abuse and DoS attacks

---

## 1. Authentication & Authorization

### Overview

MultiFish supports flexible authentication to secure API endpoints. You can choose from:
- **No Authentication** (`none`) - For development/trusted networks
- **Basic Authentication** (`basic`) - Username/password authentication
- **Token Authentication** (`token`) - Bearer token-based authentication

📖 **[Full Authentication Guide](docs/AUTHENTICATION.md)** - Comprehensive documentation with examples

### Quick Configuration

**Disable Authentication (Development)**:
```yaml
auth:
  enabled: false
  mode: none
```

**Enable Basic Authentication**:
```yaml
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "admin"
    password: "your-secure-password"
```

**Enable Token Authentication (Recommended for Production)**:
```yaml
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "your-secret-token-here"
```

### Environment Variables

```bash
# Enable authentication
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="token1,token2,token3"

# Or for basic auth
export AUTH_MODE=basic
export BASIC_AUTH_USERNAME=admin
export BASIC_AUTH_PASSWORD=secret123
```

### API Usage Examples

**With Token Auth**:
```bash
curl -H "Authorization: Bearer your-token" \
     http://localhost:8080/MultiFish/v1
```

**With Basic Auth**:
```bash
curl -u admin:password \
     http://localhost:8080/MultiFish/v1
```

### Security Notes

- ⚠️ **Certificate authentication (mTLS) is NOT supported** - We don't have a PKI infrastructure to verify client certificates
- Use HTTPS in production (via reverse proxy)
- Rotate tokens regularly
- Never commit credentials to version control
- Use environment variables for sensitive values

See **[docs/AUTHENTICATION.md](docs/AUTHENTICATION.md)** for detailed configuration, examples, and best practices.

---

## 2. Password Masking

### Implementation

A utility function has been created to consistently mask passwords across all API responses:

**Location**: `utility/security.go`

```go
func MaskPassword(password string) string {
    if password == "" {
        return ""
    }
    return "******"
}
```

### Where Password Masking is Applied

#### a) List Machines Endpoint
- **Endpoint**: `GET /MultiFish/v1/Platform`
- **File**: `handlePlatform.go`
- **Function**: `ListMachines()`
- **Protection**: All machine passwords are masked before returning the list

#### b) Get Machine Details Endpoint
- **Endpoint**: `GET /MultiFish/v1/Platform/:machineId`
- **File**: `handlePlatform.go`
- **Function**: `getMachine()`
- **Protection**: Password field in Connection object is masked

### Testing

Tests are provided in `utility/security_test.go` to ensure:
- Passwords are consistently masked
- Empty passwords are handled correctly
- The masking function is deterministic

### Best Practices

✅ **DO**:
- Always use `utility.MaskPassword()` when returning machine configurations
- Apply masking at the response layer, not data storage layer
- Test all endpoints that might expose credentials

❌ **DON'T**:
- Store masked passwords in the database/configuration
- Hardcode the mask string "******" directly
- Expose passwords in logs or error messages

---

## 3. Rate Limiting

### Implementation

A token bucket rate limiter has been implemented to protect against API abuse:

**Location**: `middleware/ratelimit.go`

### Features

- **Per-IP Rate Limiting**: Each client IP has its own rate limit
- **Configurable Limits**: Rate and burst size can be configured
- **Token Bucket Algorithm**: Allows short bursts while maintaining average rate
- **Automatic Cleanup**: Prevents memory leaks from stale entries

### Configuration

Rate limiting can be configured via:

#### 1. Configuration File (YAML)
```yaml
# config.example.yaml
rate_limit_enabled: true
rate_limit_rate: 10.0     # Requests per second (per IP)
rate_limit_burst: 20      # Maximum burst size
```

#### 2. Environment Variables
```bash
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_RATE=10.0
export RATE_LIMIT_BURST=20
```

### Default Values

- **Development** (`config.example.yaml`):
  - Rate: 10 req/s per IP
  - Burst: 20 requests
  
- **Production** (`config.production.yaml`):
  - Rate: 5 req/s per IP
  - Burst: 10 requests

### How It Works

```
1. Client makes request → 2. Extract IP address → 3. Check rate limit
                                                         ↓
                                     ← 429 Too Many Requests ← Rate exceeded?
                                                         ↓
                                     → Continue to handler ← Within limit
```

### Response on Rate Limit Exceeded

**HTTP Status**: `429 Too Many Requests`

**Response Body**:
```json
{
  "error": {
    "@Message.ExtendedInfo": [
      {
        "MessageId": "Base.1.0.RateLimitExceeded",
        "Message": "Rate limit exceeded. Please try again later.",
        "Severity": "Warning"
      }
    ]
  }
}
```

### Testing

Comprehensive tests are provided in `middleware/ratelimit_test.go`:

- ✅ Burst requests handling
- ✅ Rate limit enforcement
- ✅ Per-IP isolation
- ✅ Recovery after waiting
- ✅ Cleanup mechanism

### Monitoring

Rate limit violations are logged:

```
WARN Rate limit exceeded ip=192.168.1.100 path=/MultiFish/v1/Platform
```

### Tuning Guidelines

| Use Case | Rate (req/s) | Burst | Notes |
|----------|-------------|-------|-------|
| **Development** | 10-100 | 20-50 | Relaxed for testing |
| **Production (Internal)** | 10-20 | 20-30 | Moderate protection |
| **Production (Public)** | 1-5 | 5-10 | Strict protection |
| **High Traffic** | 50-100 | 100-200 | With caching |

### Advanced Configuration

To disable rate limiting (NOT recommended for production):

```yaml
rate_limit_enabled: false
```

Or via environment:
```bash
export RATE_LIMIT_ENABLED=false
```

### Memory Management

The rate limiter automatically cleans up stale entries when the number of tracked IPs exceeds 10,000. This prevents memory leaks in long-running deployments.

---

## 4. Additional Security Recommendations

### Implemented ✅
- [x] **Authentication & Authorization** - Basic and Token auth modes
- [x] Password masking in all API responses
- [x] Rate limiting per IP address
- [x] Configurable security settings
- [x] Comprehensive test coverage

### Future Enhancements 🚧

Consider implementing these additional security measures:

1. **Advanced Authentication**
   - JWT token authentication with expiration
   - OAuth 2.0 integration
   - Role-based access control (RBAC)
   - Per-endpoint permissions

2. **Transport Security**
   - Enforce HTTPS only
   - TLS certificate validation
   - HSTS headers

3. **Input Validation**
   - Stricter payload validation
   - SQL injection prevention
   - XSS protection

4. **Audit Logging**
   - Log all authentication attempts
   - Track configuration changes
   - Monitor suspicious activity

5. **Advanced Rate Limiting**
   - Different limits for different endpoints
   - Whitelist trusted IPs
   - Adaptive rate limiting based on behavior

6. **Secrets Management**
   - Use environment variables for sensitive data
   - Integrate with vault systems (HashiCorp Vault, AWS Secrets Manager)
   - Rotate credentials regularly

### Why Certificate Authentication is NOT Supported

**Certificate authentication (mTLS) requires:**
- Certificate Authority (CA) infrastructure
- Certificate issuance and revocation system
- Certificate validation service
- Regular certificate renewal process

**MultiFish is designed as:**
- Lightweight API service
- Simple deployment model
- No external dependencies for authentication

**Alternative for certificate-based security:**
Use a reverse proxy with mTLS support (nginx, Envoy) that validates certificates and forwards authenticated requests to MultiFish.

---

## 5. Security Testing

### Running Security Tests

```bash
# Test authentication middleware
go test ./middleware -run TestAuthMiddleware -v

# Test password masking
go test ./utility -run TestMaskPassword -v

# Test rate limiting
go test ./middleware -run TestRateLimiter -v

# Run all security-related tests
go test ./... -v | grep -E "(security|rate|mask|auth)"
```

### Manual Testing

#### Test Authentication
```bash
# Test without authentication (should fail if auth is enabled)
curl http://localhost:8080/MultiFish/v1

# Test with token authentication
curl -H "Authorization: Bearer your-token" \
     http://localhost:8080/MultiFish/v1

# Test with basic authentication
curl -u admin:password \
     http://localhost:8080/MultiFish/v1
```

#### Test Password Masking
```bash
# List machines - verify passwords are masked
curl http://localhost:8080/MultiFish/v1/Platform

# Get machine details - verify password in Connection is masked
curl http://localhost:8080/MultiFish/v1/Platform/machine1
```

#### Test Rate Limiting
```bash
# Send rapid requests to trigger rate limit
for i in {1..25}; do
  curl -w "\n%{http_code}\n" http://localhost:8080/MultiFish/v1/Platform
done

# Should see some 429 responses after burst limit
```

---

## 6. Deployment Considerations

### Production Checklist

- [ ] **Enable authentication** (`auth.enabled: true`)
- [ ] **Choose auth mode** (token recommended for production)
- [ ] **Set secure credentials** via environment variables
- [ ] Set `rate_limit_enabled: true` in production config
- [ ] Configure appropriate rate limits for your use case
- [ ] **Use HTTPS** (configure reverse proxy)
- [ ] Monitor rate limit violations in logs
- [ ] Monitor authentication failures in logs
- [ ] Verify password masking on all endpoints
- [ ] Set up alerts for excessive 429 responses
- [ ] Document rate limits and auth requirements in API docs
- [ ] Consider CDN/reverse proxy for additional protection

### Environment Variables

```bash
# Recommended production settings
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="$(openssl rand -hex 32)"
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_RATE=5.0
export RATE_LIMIT_BURST=10
export LOG_LEVEL=info
```

---

## 7. Troubleshooting

### Authentication Issues

**Problem**: Getting 401 Unauthorized errors

**Solutions**:
- Verify authentication is enabled in config
- Check credentials format (Bearer token or Basic auth)
- Review logs for authentication attempts
- Test with curl verbose mode: `curl -v ...`

**Problem**: Authentication not enforced

**Check**:
- Verify `auth.enabled: true` in config
- Confirm `auth.mode` is not "none"
- Check middleware is registered in `main.go`
- Look for "Authentication enabled" in logs

### Rate Limiting Issues

**Problem**: Legitimate users getting rate limited

**Solutions**:
- Increase `rate_limit_rate` or `rate_limit_burst`
- Implement IP whitelisting for trusted sources
- Use a reverse proxy to aggregate requests

**Problem**: Rate limiting not working

**Check**:
- Verify `rate_limit_enabled: true` in config
- Check logs for rate limiter initialization
- Confirm middleware is registered in `main.go`

### Password Masking Issues

**Problem**: Passwords visible in responses

**Solutions**:
- Verify `utility.MaskPassword()` is called
- Check for direct struct serialization bypassing masking
- Review custom response builders

---

## 8. References

- [Authentication Guide](docs/AUTHENTICATION.md) - Comprehensive authentication documentation
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [Redfish Specification](https://www.dmtf.org/standards/redfish)

---

**Last Updated**: 2026-02-10  
**Version**: 2.0.0

# MultiFish Security Guide

This document is the single source of truth for MultiFish security configuration and operations.

## Overview

MultiFish security focuses on three core controls:

1. Authentication & authorization
2. Password masking in API responses
3. Per-IP rate limiting

---

## 1) Authentication & Authorization

### Supported Modes

- `none`: No authentication (development only)
- `basic`: HTTP Basic authentication
- `token`: Bearer token authentication (recommended for production)

### Quick Start

Disable auth (development):

```yaml
auth:
  enabled: false
  mode: none
```

Enable basic auth:

```yaml
auth:
  enabled: true
  mode: basic
  basic_auth:
    username: "admin"
    password: "your-secure-password"
```

Enable token auth:

```yaml
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "token-1"
      - "token-2"
```

### Environment Variables

```bash
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="token1,token2,token3"

# Basic auth alternative
# export AUTH_MODE=basic
# export BASIC_AUTH_USERNAME=admin
# export BASIC_AUTH_PASSWORD=secret123
```

### Configuration Priority

Configuration is loaded in this order:

1. Environment variables
2. YAML configuration file
3. Built-in defaults

### API Usage Examples

Token auth:

```bash
curl -H "Authorization: Bearer your-token" \
  http://localhost:8080/MultiFish/v1
```

Basic auth:

```bash
curl -u admin:password \
  http://localhost:8080/MultiFish/v1
```

### Deployment Examples

Systemd:

```ini
[Service]
Environment="AUTH_ENABLED=true"
Environment="AUTH_MODE=token"
Environment="TOKEN_AUTH_TOKENS=your-production-token"
```

Docker:

```bash
docker run -d \
  -p 8080:8080 \
  -e AUTH_ENABLED=true \
  -e AUTH_MODE=token \
  -e TOKEN_AUTH_TOKENS=docker-token-123 \
  multifish:latest
```

Kubernetes (manifests default to disabled auth; enable by command):

```bash
TOKEN1=$(openssl rand -hex 32)
TOKEN2=$(openssl rand -hex 32)

kubectl -n default create secret generic multifish-secret \
  --from-literal=auth-tokens="$TOKEN1,$TOKEN2" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl -n default set env deployment/multifish \
  AUTH_ENABLED=true \
  AUTH_MODE=token \
  TOKEN_AUTH_TOKENS="$TOKEN1,$TOKEN2"
```

### Authentication Best Practices

- Enable authentication in production
- Use token mode for production workloads
- Use environment variables / secret manager for secrets
- Use HTTPS via reverse proxy
- Rotate tokens regularly
- Use different tokens per client/integration

Generate secure token:

```bash
openssl rand -hex 32
```

### Authentication Troubleshooting

If you receive `401 Unauthorized`:

1. Verify auth mode and enabled flag
2. Verify header format
   - Basic: `Authorization: Basic <base64(username:password)>`
   - Token: `Authorization: Bearer <token>`
3. Verify token/credential matches runtime config
4. Check service logs

Useful checks:

```bash
curl -v -H "Authorization: Bearer your-token" \
  http://localhost:8080/MultiFish/v1
```

```bash
journalctl -u multifish -n 100
```

### Authentication Limitations

Current authentication model does **not** include:

- mTLS / certificate authentication
- OAuth2 / OpenID Connect
- JWT expiration/refresh flow
- RBAC (all authenticated users share same access scope)
- Per-user rate limits (rate limit is per IP)

For certificate-based client auth, use reverse proxy mTLS (nginx/Envoy) in front of MultiFish.

---

## 2) Password Masking

### Implementation

Passwords are masked before API responses are returned.

- File: `utility/security.go`
- Function: `MaskPassword(password string) string`

### Protected Endpoints

- `GET /MultiFish/v1/Platform`
- `GET /MultiFish/v1/Platform/:machineId`

### Validation

- Tests: `utility/security_test.go`
- Ensures deterministic masking and safe handling of empty passwords

### Best Practices

Do:
- Always mask response-layer secrets
- Keep raw secrets only where operationally required

Don't:
- Log credentials/tokens
- Store masked values as canonical data

---

## 3) Rate Limiting

### Implementation

Token-bucket, per-IP rate limiting middleware.

- File: `middleware/ratelimit.go`

### Configuration

YAML:

```yaml
rate_limit_enabled: true
rate_limit_rate: 10.0
rate_limit_burst: 20
```

Environment variables:

```bash
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_RATE=10.0
export RATE_LIMIT_BURST=20
```

### Behavior

When exceeded:

- HTTP status: `429 Too Many Requests`
- Error payload uses Redfish-style extended info

### Monitoring and Tuning

- Monitor warning logs for repeated 429 events
- Tune by traffic profile (internal/public/high-traffic)
- Keep cleanup behavior enabled to avoid stale-IP memory growth

---

## 4) Operational Checklist

Before production rollout:

- [ ] `auth.enabled: true`
- [ ] `auth.mode: token`
- [ ] secure tokens from environment/secret manager
- [ ] `rate_limit_enabled: true`
- [ ] HTTPS via reverse proxy
- [ ] verify masked passwords in platform APIs
- [ ] monitor auth failures and 429 spikes

Recommended baseline:

```bash
export AUTH_ENABLED=true
export AUTH_MODE=token
export TOKEN_AUTH_TOKENS="$(openssl rand -hex 32)"
export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_RATE=5.0
export RATE_LIMIT_BURST=10
export LOG_LEVEL=info
```

---

## 5) Security Testing

```bash
# Authentication
go test ./middleware -run TestAuthMiddleware -v

# Password masking
go test ./utility -run TestMaskPassword -v

# Rate limiting
go test ./middleware -run TestRateLimiter -v
```

Manual checks:

```bash
# Without token (expect 401 when auth enabled)
curl -i http://localhost:8080/MultiFish/v1

# With token (expect 200)
curl -i -H "Authorization: Bearer your-token" http://localhost:8080/MultiFish/v1
```

---

## References

- [Main README](README.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Config Guide](config/README.md)
- [OWASP API Security Top 10](https://owasp.org/www-project-api-security/)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)

# MultiFish Deployment Guide

This guide covers different deployment scenarios for MultiFish, including configuration management, systemd service setup, and best practices.

## Table of Contents

- [Configuration Management](#configuration-management)
- [Running MultiFish](#running-multifish)
- [Systemd Service Setup](#systemd-service-setup)
- [Production Deployment](#production-deployment)
- [Troubleshooting](#troubleshooting)

## Configuration Management

MultiFish supports flexible configuration management with multiple options:

### Configuration Priority

MultiFish follows this priority order for configuration:

1. **Explicit `-config` flag**: Highest priority
2. **Environment variable**: `MULTIFISH_CONFIG`
3. **Default `config.yaml`**: In the working directory
4. **Built-in defaults**: Fallback if no config file exists

### Configuration Files

#### Development Configuration
```yaml
# config.yaml
port: 8080
log_level: debug
worker_pool_size: 50
logs_dir: ./logs
shutdown_timeout: 30
rate_limit_enabled: true
rate_limit_rate: 10.0
rate_limit_burst: 20
auth:
  enabled: false
  mode: none
```

#### Production Configuration
```yaml
# config.production.yaml
port: 8080
log_level: info
worker_pool_size: 100
logs_dir: /var/log/multifish
shutdown_timeout: 60
rate_limit_enabled: true
rate_limit_rate: 100.0
rate_limit_burst: 200
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "your-production-token-here"
```

### Using Configuration Files

#### Option 1: Default Config
Place `config.yaml` in the same directory as the binary:
```bash
# Creates or edits config.yaml
cp config.example.yaml config.yaml
nano config.yaml
```

#### Option 2: Explicit Config Path
Specify the config file explicitly:
```bash
./multifish -config /path/to/config.production.yaml
```

#### Option 3: Environment Variable
Set the config path via environment variable:
```bash
export MULTIFISH_CONFIG=/path/to/config.production.yaml
./multifish
```

#### Option 4: No Config (Defaults)
Run without any configuration file:
```bash
# Ensure no default config.yaml exists
rm -f config.yaml
./multifish
```

## Running MultiFish

### Using the Management Script

The `multifish.sh` script provides convenient management commands:

#### Basic Usage
```bash
# Build the binary
./multifish.sh build

# Start with default config (config.yaml if exists)
./multifish.sh start

# Start with specific config
./multifish.sh -c config.production.yaml start

# Start without config (remove config.yaml first)
rm config.yaml
./multifish.sh start

# Check status
./multifish.sh status

# View logs
./multifish.sh logs

# Stop the service
./multifish.sh stop

# Restart
./multifish.sh restart

# Test API
./multifish.sh test
```

#### Using Environment Variables
```bash
# Set config via environment variable
export MULTIFISH_CONFIG=/path/to/config.production.yaml
./multifish.sh start

# Or use inline
MULTIFISH_CONFIG=/path/to/config.production.yaml ./multifish.sh start
```

### Direct Binary Execution

Run the binary directly for more control:

```bash
# With default config.yaml
./multifish

# With specific config
./multifish -config config.production.yaml

# Without config (built-in defaults)
./multifish

# Run in background
nohup ./multifish -config config.production.yaml > multifish.log 2>&1 &
```

## Systemd Service Setup

For production deployments, use systemd for process management.

### 1. Update Service File

Edit `multifish.service` to match your environment:

```ini
[Unit]
Description=MultiFish BMC Management API
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=yujen
Group=yujen
WorkingDirectory=/home/yujen/MultiFish

# Choose one of these options:

# Option A: With default config.yaml (recommended for most deployments)
ExecStart=/home/yujen/MultiFish/multifish

# Option B: With specific config file
# ExecStart=/home/yujen/MultiFish/multifish -config /home/yujen/MultiFish/config.production.yaml

# Option C: Via environment variable
# Environment="MULTIFISH_CONFIG=/home/yujen/MultiFish/config.production.yaml"
# ExecStart=/home/yujen/MultiFish/multifish

Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

### 2. Install Service

```bash
# Copy service file
sudo cp multifish.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable on boot (optional)
sudo systemctl enable multifish

# Start service
sudo systemctl start multifish

# Check status
sudo systemctl status multifish
```

### 3. Service Management

```bash
# Start service
sudo systemctl start multifish

# Stop service
sudo systemctl stop multifish

# Restart service
sudo systemctl restart multifish

# Reload configuration (if supported)
sudo systemctl reload multifish

# Enable on boot
sudo systemctl enable multifish

# Disable on boot
sudo systemctl disable multifish

# View logs
sudo journalctl -u multifish -f

# View last 100 lines
sudo journalctl -u multifish -n 100

# View logs since boot
sudo journalctl -u multifish -b
```

## Docker Deployment

Docker provides an easy way to containerize and deploy MultiFish with all its dependencies.

### Prerequisites

- Docker 20.10+ installed
- Basic understanding of Docker concepts

### Quick Start with Docker

#### 1. Create Dockerfile

Create a `Dockerfile` in your project root:

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o multifish

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections to BMCs
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/multifish .

# Copy configuration files (optional)
COPY config.example.yaml ./config.yaml

# Create logs directory
RUN mkdir -p /app/logs

# Expose port
EXPOSE 8080

# Run the application
CMD ["./multifish"]
```

#### 2. Build Docker Image

```bash
# Build image
docker build -t multifish:latest .

# Build with specific tag
docker build -t multifish:v1.0.0 .

# Build with no cache
docker build --no-cache -t multifish:latest .
```

#### 3. Run Container

**Option A: Without configuration file (use defaults)**
```bash
docker run -d \
  --name multifish \
  -p 8080:8080 \
  -e LOG_LEVEL=info \
  -e WORKER_POOL_SIZE=100 \
  multifish:latest
```

**Option B: With configuration file**
```bash
# Create config file first
cat > config.yaml << EOF
port: 8080
log_level: info
worker_pool_size: 100
logs_dir: /app/logs
shutdown_timeout: 60
rate_limit_enabled: true
rate_limit_rate: 100.0
rate_limit_burst: 200
auth:
  enabled: true
  mode: token
  token_auth:
    tokens:
      - "your-production-token-here"
EOF

# Run with mounted config
docker run -d \
  --name multifish \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/logs:/app/logs \
  multifish:latest ./multifish -config /app/config.yaml
```

**Option C: With environment variables (recommended)**
```bash
docker run -d \
  --name multifish \
  -p 8080:8080 \
  -e PORT=8080 \
  -e LOG_LEVEL=info \
  -e WORKER_POOL_SIZE=100 \
  -e LOGS_DIR=/app/logs \
  -e AUTH_ENABLED=true \
  -e AUTH_MODE=token \
  -e TOKEN_AUTH_TOKENS="token1,token2,token3" \
  -v $(pwd)/logs:/app/logs \
  multifish:latest
```

#### 4. Container Management

```bash
# Check status
docker ps

# View logs
docker logs multifish
docker logs -f multifish  # Follow mode

# Stop container
docker stop multifish

# Start container
docker start multifish

# Restart container
docker restart multifish

# Remove container
docker rm -f multifish

# Execute commands in container
docker exec -it multifish /bin/sh

# Inspect container
docker inspect multifish
```

### Docker Troubleshooting

#### Container Won't Start
```bash
# Check logs
docker logs multifish

# Check container status
docker ps -a

# Inspect container
docker inspect multifish

# Check resource usage
docker stats multifish
```

#### Configuration Issues
```bash
# Verify environment variables
docker exec multifish env

# Check mounted files
docker exec multifish ls -la /app/config.yaml

# Test configuration
docker exec multifish cat /app/config.yaml
```

#### Network Issues
```bash
# Check network
docker network ls
docker network inspect multifish-network

# Test connectivity
docker exec multifish wget -O- http://localhost:8080/MultiFish/v1
```

#### Permission Issues
```bash
# Check file permissions
docker exec multifish ls -la /app/logs

# Run as specific user
docker run --user 1000:1000 ...
```

## Kubernetes Deployment

Kubernetes provides production-grade orchestration for containerized applications.

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Basic knowledge of Kubernetes concepts

### 1. ConfigMap

Create `k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: multifish-config
  namespace: default
data:
  config.yaml: |
    port: 8080
    log_level: info
    worker_pool_size: 100
    logs_dir: /var/log/multifish
    shutdown_timeout: 60
    rate_limit_enabled: true
    rate_limit_rate: 100.0
    rate_limit_burst: 200
    auth:
      enabled: true
      mode: token
```

### 2. Secret

Create `k8s/secret.yaml`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: multifish-secret
  namespace: default
type: Opaque
stringData:
  # Token authentication
  auth-tokens: |
    your-secret-token-1
    your-secret-token-2
    your-secret-token-3
```

Create secret from command line:
```bash
# From literal values
kubectl create secret generic multifish-secret \
  --from-literal=auth-tokens="token1,token2,token3"

# From file
kubectl create secret generic multifish-secret \
  --from-file=auth-tokens=./tokens.txt
```

### 3. Deployment

Create `k8s/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: multifish
  namespace: default
  labels:
    app: multifish
spec:
  replicas: 2
  selector:
    matchLabels:
      app: multifish
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    metadata:
      labels:
        app: multifish
    spec:
      containers:
      - name: multifish
        image: multifish:latest
        imagePullPolicy: Always
        
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
        
        # Environment variables from ConfigMap
        env:
        - name: PORT
          value: "8080"
        - name: LOG_LEVEL
          valueFrom:
            configMapKeyRef:
              name: multifish-config
              key: log_level
        - name: WORKER_POOL_SIZE
          value: "100"
        - name: LOGS_DIR
          value: "/var/log/multifish"
        - name: AUTH_ENABLED
          value: "true"
        - name: AUTH_MODE
          value: "token"
        - name: TOKEN_AUTH_TOKENS
          valueFrom:
            secretKeyRef:
              name: multifish-secret
              key: auth-tokens
        
        # Volume mounts
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
          readOnly: true
        - name: logs
          mountPath: /var/log/multifish
        
        # Command
        command: ["./multifish"]
        args: ["-config", "/app/config.yaml"]
        
        # Resource limits
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
        
        # Liveness probe
        livenessProbe:
          httpGet:
            path: /MultiFish/v1
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        
        # Readiness probe
        readinessProbe:
          httpGet:
            path: /MultiFish/v1
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
      
      # Volumes
      volumes:
      - name: config
        configMap:
          name: multifish-config
      - name: logs
        emptyDir: {}
      
      # Security context
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
```

### 4. Service

Create `k8s/service.yaml`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: multifish-service
  namespace: default
  labels:
    app: multifish
spec:
  type: ClusterIP
  selector:
    app: multifish
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800
```

### 5. Ingress (Optional)

Create `k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: multifish-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - multifish.example.com
    secretName: multifish-tls
  rules:
  - host: multifish.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: multifish-service
            port:
              number: 8080
```

### 6. PersistentVolumeClaim (Optional)

Create `k8s/pvc.yaml` for persistent logs:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: multifish-logs-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
```

Update deployment to use PVC:
```yaml
      volumes:
      - name: logs
        persistentVolumeClaim:
          claimName: multifish-logs-pvc
```

### Kubernetes Deployment Commands

```bash
# If you don't have a cluster running or configured locally (like minikube, kind, or a remote kubeconfig), then do the following processes first.
# Option 1: Start a Local Cluster with minikube
# If you're working locally (e.g., in Codespaces or a dev environment):
# 1. **Install Minikube** if you haven’t:
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
# 2. **Start a cluster**:
minikube start
minikube stop --all
# Option 2: Use a Cloud Cluster (e.g., GKE, EKS, AKS)**
# If you're deploying to a cloud provider, make sure:
# * You have the correct `kubeconfig` set up (`~/.kube/config`)
# * You're authenticated (e.g., via `gcloud`, `aws`, or `az`)
# * Then try:
kubectl config use-context <your-cluster-context>
kubectl apply -f ./k8s

# Create namespace (optional)
kubectl create namespace multifish

# Apply configurations
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Or apply all at once
kubectl apply -f k8s/

# Check deployment status
kubectl get deployments
kubectl get pods
kubectl get services

# View logs
kubectl logs -f deployment/multifish

# View logs from specific pod
kubectl logs -f pod/multifish-xxxxx

# Execute commands in pod
kubectl exec -it deployment/multifish -- /bin/sh

# Port forward for local testing
kubectl port-forward service/multifish-service 8080:8080

# Scale deployment
kubectl scale deployment multifish --replicas=3

# Update image
kubectl set image deployment/multifish multifish=multifish:v1.0.1

# Rollback deployment
kubectl rollout undo deployment/multifish

# Check rollout status
kubectl rollout status deployment/multifish

# Delete resources
kubectl delete -f k8s/
```

### Enable Authentication (Token Mode) via Commands

Use this flow when your current Kubernetes deployment is running with auth disabled and you want to turn on token auth from CLI.

1. Generate secure tokens:
```bash
TOKEN1=$(openssl rand -hex 32)
TOKEN2=$(openssl rand -hex 32)
echo "Generated tokens:" && echo "$TOKEN1" && echo "$TOKEN2"
```

2. Create or update Secret (idempotent):
```bash
kubectl -n default create secret generic multifish-secret \
  --from-literal=auth-tokens="$TOKEN1,$TOKEN2" \
  --dry-run=client -o yaml | kubectl apply -f -
```

3. Enable token auth on Deployment:
```bash
kubectl -n default set env deployment/multifish \
  AUTH_ENABLED=true \
  AUTH_MODE=token \
  TOKEN_AUTH_TOKENS="$TOKEN1,$TOKEN2"
```

4. Verify rollout and test:
```bash
kubectl -n default rollout status deployment/multifish
kubectl -n default get pods -l app=multifish

# Port-forward for local test
kubectl -n default port-forward service/multifish-service 8080:8080

# Without token (should be rejected)
curl -i http://127.0.0.1:8080/MultiFish/v1

# With token (should succeed)
curl -i -H "Authorization: Bearer $TOKEN1" http://127.0.0.1:8080/MultiFish/v1
```

Important:
- `kubectl create secret generic ... --from-literal=auth-tokens=...` is correct.
- Secret creation alone does not enable auth.
- Auth is active only when Deployment runtime config has `AUTH_ENABLED=true` and `AUTH_MODE=token`.

### Kubernetes Best Practices

1. **Use ConfigMaps and Secrets**
   - ConfigMaps for non-sensitive configuration
   - Secrets for sensitive data (tokens, passwords)

2. **Resource Limits**
   - Always set resource requests and limits
   - Prevents resource exhaustion

3. **Health Checks**
   - Liveness probe: Restart unhealthy containers
   - Readiness probe: Control traffic routing

4. **Multiple Replicas**
   - Run at least 2 replicas for high availability
   - Use RollingUpdate strategy

5. **Monitoring and Logging**
   - Use centralized logging (ELK, Loki)
   - Implement metrics (Prometheus)

## Production Deployment

### Best Practices

#### 1. Use Production Configuration
```bash
# Create production config
cp config.example.yaml config.production.yaml
nano config.production.yaml
```

Key settings for production:
- Set `log_level: info` or `warn`
- Enable authentication: `auth.enabled: true`
- Enable rate limiting: `rate_limit_enabled: true`
- Increase `worker_pool_size` for high concurrency
- Set appropriate `shutdown_timeout`
- Configure proper `logs_dir` (e.g., `/var/log/multifish`)

#### 2. Secure Configuration Files
```bash
# Set proper permissions
chmod 600 config.production.yaml
chown yujen:yujen config.production.yaml

# Store sensitive configs outside web root
mv config.production.yaml /etc/multifish/config.yaml
chmod 600 /etc/multifish/config.yaml
```

#### 3. Log Management
```bash
# Create logs directory
sudo mkdir -p /var/log/multifish
sudo chown yujen:yujen /var/log/multifish

# Setup log rotation
sudo nano /etc/logrotate.d/multifish
```

Logrotate configuration:
```
/var/log/multifish/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 yujen yujen
    sharedscripts
    postrotate
        systemctl reload multifish > /dev/null 2>&1 || true
    endscript
}
```

#### 4. Firewall Configuration
```bash
# Allow MultiFish port (default 8080)
sudo ufw allow 8080/tcp
sudo ufw status
```

#### 5. Reverse Proxy Setup (Optional)

For production, consider using nginx as a reverse proxy:

```nginx
# /etc/nginx/sites-available/multifish
server {
    listen 80;
    server_name multifish.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable and restart nginx:
```bash
sudo ln -s /etc/nginx/sites-available/multifish /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### Deployment Checklist

- [ ] Build optimized binary: `go build -ldflags="-s -w" -o multifish`
- [ ] Create production config file
- [ ] Enable authentication and rate limiting
- [ ] Set proper file permissions
- [ ] Configure logs directory
- [ ] Setup log rotation
- [ ] Install systemd service
- [ ] Configure firewall
- [ ] Test service startup and restart
- [ ] Setup monitoring/alerting
- [ ] Configure reverse proxy (if needed)
- [ ] Setup SSL/TLS (if needed)

## Troubleshooting

### Service Won't Start

1. **Check service status**:
```bash
sudo systemctl status multifish
sudo journalctl -u multifish -n 50
```

2. **Verify binary path**:
```bash
ls -la /home/yujen/MultiFish/multifish
```

3. **Test binary manually**:
```bash
cd /home/yujen/MultiFish
./multifish
# Or with config
./multifish -config config.yaml
```

4. **Check configuration file**:
```bash
# Verify config file exists and is readable
cat config.yaml

# Test with explicit config
./multifish -config config.production.yaml
```

5. **Verify permissions**:
```bash
# Binary should be executable
chmod +x /home/yujen/MultiFish/multifish

# Config should be readable
chmod 644 config.yaml

# Logs directory should be writable
mkdir -p logs
chmod 755 logs
```

### Config File Not Found

If you see "Failed to load configuration" errors:

1. **Check default config**:
```bash
ls -la config.yaml
```

2. **Specify config explicitly**:
```bash
./multifish -config /full/path/to/config.yaml
```

3. **Run without config** (uses defaults):
```bash
# Temporarily rename or remove config.yaml
mv config.yaml config.yaml.bak
./multifish
```

### Port Already in Use

```bash
# Check what's using the port
sudo lsof -i :8080

# Kill the process
sudo kill -9 <PID>

# Or change port in config.yaml
nano config.yaml
# Set: port: 8081
```

### Permission Denied Errors

```bash
# Ensure logs directory exists and is writable
mkdir -p logs
chmod 755 logs

# If using /var/log/multifish
sudo mkdir -p /var/log/multifish
sudo chown yujen:yujen /var/log/multifish
```

### Service Stops Unexpectedly

1. **Check logs**:
```bash
sudo journalctl -u multifish -n 100
tail -f /var/log/multifish/multifish.log
```

2. **Increase restart limits** in service file:
```ini
StartLimitBurst=5
StartLimitIntervalSec=120s
```

3. **Add resource limits**:
```ini
MemoryMax=2G
LimitNOFILE=65536
```

### Configuration Not Applied

1. **Verify config is being loaded**:
```bash
# Check startup logs
journalctl -u multifish -n 50 | grep "Configuration loaded"
```

2. **Ensure correct config path**:
```bash
# In systemd service
ExecStart=/home/yujen/MultiFish/multifish -config /home/yujen/MultiFish/config.production.yaml
```

3. **Restart after config changes**:
```bash
sudo systemctl restart multifish
```

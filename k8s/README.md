# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying MultiFish to a Kubernetes cluster.

## Files

- `configmap.yaml` - Application configuration
- `secret.yaml` - Sensitive data (tokens, passwords)
- `deployment.yaml` - Main application deployment
- `service.yaml` - Service for internal communication
- `ingress.yaml` - External access configuration (optional)

## Quick Start

### 1. Update Configuration

Edit the files to match your environment:

**secret.yaml:**
```bash
# Update with your actual tokens
kubectl create secret generic multifish-secret \
  --from-literal=auth-tokens="your-actual-token-1,your-actual-token-2"
```

**ingress.yaml:**
```yaml
# Update domain name
- host: multifish.your-domain.com
```

### 2. Deploy to Kubernetes

```bash
# Create namespace (optional)
kubectl create namespace multifish

# Apply all configurations
kubectl apply -f k8s/

# Or apply individually
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

### 3. Verify Deployment

```bash
# Check deployment status
kubectl get deployments
kubectl get pods
kubectl get services

# View logs
kubectl logs -f deployment/multifish

# Check health
kubectl get pods -l app=multifish
```

### 4. Access Application

```bash
# Port forward for local access
kubectl port-forward service/multifish-service 8080:8080

# Then access at http://localhost:8080/MultiFish/v1
```

## Management Commands

```bash
# Scale deployment
kubectl scale deployment multifish --replicas=3

# Update image
kubectl set image deployment/multifish multifish=multifish:v1.0.1

# Rollback
kubectl rollout undo deployment/multifish

# View rollout history
kubectl rollout history deployment/multifish

# Delete deployment
kubectl delete -f k8s/
```

## Customization

### Using Different Namespace

Add to all manifests:
```yaml
metadata:
  namespace: your-namespace
```

### Persistent Logs

Create PVC (pvc.yaml):
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: multifish-logs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

Update deployment:
```yaml
volumes:
- name: logs
  persistentVolumeClaim:
    claimName: multifish-logs-pvc
```

### External Load Balancer

Change service type:
```yaml
spec:
  type: LoadBalancer
```

## Troubleshooting

See [DEPLOYMENT.md](../DEPLOYMENT.md#kubernetes-deployment) for detailed troubleshooting guide.

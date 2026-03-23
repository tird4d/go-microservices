# 🚀 Microservice Deployment & Production Guide

This document defines the full checklist and best practices for deploying Go-based microservices into a production-grade Kubernetes environment.
It is designed to be reusable for any service in your system.

---

## 📋 1. Build Production-Ready Docker Images

- Use multi-stage builds.
- Only copy final binary to small runtime image (e.g., `distroless` or `alpine`).
- Set environment variables correctly (`ENV` in Dockerfile).

**Example Dockerfile:**
```Dockerfile
FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod . 
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o main .

FROM gcr.io/distroless/static
COPY --from=builder /app/main /
CMD ["/main"]
```

**Build:**
```bash
docker build -t your-service:latest .
```

**Push to Registry:** (optional if using remote cluster)
```bash
docker push your-service:latest
```

---

## 📋 2. Helm Charts (Production Values)

- Create separate `values-prod.yaml`.
- Configure:
  - Resource limits (CPU, Memory)
  - LivenessProbe & ReadinessProbe
  - Metrics scraping annotations
  - Replicas > 1 (at least 2)
  - Secrets management

**Example:**
```yaml
replicaCount: 2
resources:
  limits:
    cpu: "500m"
    memory: "512Mi"
  requests:
    cpu: "200m"
    memory: "256Mi"
livenessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:50051"]
readinessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:50051"]
```

Deploy:
```bash
helm upgrade --install your-service ./charts/your-service -f values-prod.yaml
```

---

## 📋 3. Set Up Kubernetes Namespaces

- Create dedicated namespaces:
  - `dev`
  - `staging`
  - `prod`

**Example:**
```bash
kubectl create namespace prod
kubectl create namespace dev
kubectl create namespace staging
```

Use namespace in Helm:
```bash
helm upgrade --install your-service ./charts/your-service --namespace prod
```

---

## 📋 4. Ingress Controller (Expose Services)

- Install NGINX Ingress Controller:
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install nginx-ingress ingress-nginx/ingress-nginx
```

- Create Ingress resource for your services.

**Example ingress.yaml:**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: your-service-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
spec:
  rules:
  - host: your-service.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: your-service
            port:
              number: 50051
```

Apply:
```bash
kubectl apply -f ingress.yaml
```

---

## 📋 5. Enable TLS (HTTPS)

- Use cert-manager to issue certificates automatically.

Install cert-manager:
```bash
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager --set installCRDs=true
```

Create Certificate resource.

---

## 📋 6. Monitoring and Alerting (Production Level)

- Install Prometheus Operator.
- Monitor:
  - CPU usage
  - Memory usage
  - Request count per service
  - Error rates
- Add Grafana Dashboards for each service.
- Configure Alertmanager for critical alerts (Slack, Email, etc).

---

## 📋 7. CI/CD (Later Stage)

- Build Docker Image automatically
- Push to Registry
- Helm Upgrade Deployment via GitHub Actions or GitLab CI

**Example GitHub Action:**
```yaml
name: Deploy to Prod
on:
  push:
    branches:
      - main
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - name: Checkout
      uses: actions/checkout@v2
    - name: Build Docker image
      run: docker build -t your-service .
    - name: Push to registry
      run: docker push your-service
    - name: Helm Upgrade
      run: helm upgrade --install your-service ./charts/your-service --namespace prod -f values-prod.yaml
```

---

# 🎯 Done!
You now have a clear, production-ready deployment path for your Go microservices!
Use this checklist every time you launch a new service 🚀

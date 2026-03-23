# 🚀 Full Setup Guide: Deploying a Go Microservice with Kubernetes, Monitoring, and Dependencies

This guide provides a full A-to-Z setup for launching a Go-based microservice using Kubernetes, with everything included:
- Install Minikube & Helm
- Setup Redis, MongoDB, RabbitMQ
- Define gRPC services using Protobuf
- Build and containerize the service
- Write and apply Helm charts
- Add observability: Prometheus + Grafana + Alerting
- gRPC health checks (readiness & liveness)

---

## 📦 Prerequisites
- [ ] Minikube installed: https://minikube.sigs.k8s.io/docs/start/
- [ ] Helm installed: https://helm.sh/docs/intro/install/
- [ ] Docker installed and running
- [ ] Go 1.20+ installed
- [ ] kubectl installed

---

## 🧱 Step 1: Start Kubernetes Locally (Minikube)
```bash
minikube start
```
Enable Docker build context inside Minikube:
```bash
eval $(minikube docker-env)
```

---

## 🧰 Step 2: Install Helm Charts for Dependencies

### Add Bitnami Repo
```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update
```

### Install MongoDB
```bash
helm install mongo-release bitnami/mongodb
```

### Install Redis
```bash
helm install redis-release bitnami/redis   --set auth.enabled=false   --set architecture=standalone
```

### Install RabbitMQ
```bash
helm install rabbitmq bitnami/rabbitmq   --set auth.username=admin   --set auth.password=adminpassword   --set auth.erlangCookie=supersecretcookie
```

---

## 📜 Step 3: Generate gRPC Files with Protobuf

Install Protobuf compiler (if needed): https://grpc.io/docs/protoc-installation/

Install Go plugins:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `$GOPATH/bin` is in your PATH. Then generate files:
```bash
protoc --go_out=. --go-grpc_out=. proto/user.proto

#more parameter
protoc \
    --go_out=. \
    --go-grpc_out=.  \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    proto/auth.proto
```

---

## 🐳 Step 4: Build Docker Image for Microservice

```bash
docker build -t user-service:latest ./user_service
minikube image load user-service:latest
```

---

## 🧱 Step 5: Create Helm Chart for Microservice

Generate Helm chart:
```bash
helm create user-service
```

Edit `values.yaml`:
```yaml
image:
  repository: user-service
  tag: latest

service:
  port: 50051
  metricsPort: 9090

readinessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:50051"]
  initialDelaySeconds: 5
  periodSeconds: 10

livenessProbe:
  exec:
    command: ["/bin/grpc-health-probe", "-addr=:50051"]
  initialDelaySeconds: 5
  periodSeconds: 10

env:
  - name: RABBITMQ_URL
    value: amqp://admin:adminpassword@rabbitmq.default.svc.cluster.local:5672/
  - name: MONGO_URI
    value: mongodb://mongo-release-mongodb.default.svc.cluster.local:27017
```

Set up `deployment.yaml`, `service.yaml`, add probes, annotations, and port mappings.

Deploy:
```bash
helm upgrade --install user-service ./charts/user-service
```

---

## 🔍 Step 6: Add Logging & Metrics

### Structured Logging (zap)
In `logger/logger.go`:
```go
var Log *zap.SugaredLogger
func InitLogger(debug bool) { ... }
```

### Prometheus Metrics
- Use `prometheus.NewCounterVec`, `NewHistogramVec`
- Create `/metrics` HTTP endpoint (e.g., on port `9090`)

In service:
```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/path: /metrics
  prometheus.io/port: "9090"
```

---

## 📈 Step 7: Monitoring with Prometheus + Grafana

### Install Prometheus
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/prometheus
```

### Install Grafana
```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install grafana grafana/grafana
```

Get password and port-forward:
```bash
kubectl get secret grafana -o jsonpath="{.data.admin-password}" | base64 --decode
kubectl port-forward deployment/grafana 3000:3000
```
Visit: `http://localhost:3000`
Add Prometheus datasource:
```
http://prometheus-server.default.svc.cluster.local
```

---

## 🚨 Step 8: Add Alerting Rules

Create `prometheus-values.yaml`:
```yaml
alertmanager:
  enabled: true

serverFiles:
  alerting_rules.yml:
    groups:
      - name: user-service-alerts
        rules:
          - alert: HighRegisterUserRequestRate
            expr: increase(user_service_requests_total{endpoint="RegisterUser"}[5m]) > 10
            for: 1m
            labels:
              severity: warning
            annotations:
              summary: "High rate of RegisterUser requests detected"
              description: "RegisterUser endpoint received more than 10 requests in the last 5 minutes."
```

Apply:
```bash
helm upgrade prometheus prometheus-community/prometheus -f prometheus-values.yaml
```

Check: `http://localhost:9090/alerts`

---

## ✅ Final Checklist
- [x] Kubernetes running via Minikube
- [x] Redis, MongoDB, RabbitMQ installed
- [x] Protobuf generated for gRPC
- [x] Microservice built and deployed
- [x] Helm chart configured and applied
- [x] Logging and Prometheus metrics integrated
- [x] Prometheus and Grafana installed
- [x] Alerting rules firing and visible in UI
- [x] gRPC readiness and liveness probes set with grpc-health-probe

---

## 🎯 Done!
You now have a complete blueprint to bootstrap any new Go microservice in a cloud-native, observable, production-ready architecture.
Use this file as your tutorial and repeatable setup guide.

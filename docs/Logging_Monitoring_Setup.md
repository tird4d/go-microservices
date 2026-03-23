
# 📅 Setup Guide: Logging and Monitoring for Go Microservices (User-Service)

This document covers the full process to set up **structured logging**, **Prometheus metrics**, **Prometheus server**, and **Grafana dashboard** for a Go microservice inside a Kubernetes cluster.

---

## ✨ 1. Structured Logging (zap)

### Install zap library
```bash
go get go.uber.org/zap
```

### Create a logger utility
```go
package logger

import "go.uber.org/zap"

var Log *zap.SugaredLogger

func InitLogger(debug bool) {
    var logger *zap.Logger
    if debug {
        logger, _ = zap.NewDevelopment()
    } else {
        logger, _ = zap.NewProduction()
    }
    Log = logger.Sugar()
}
```

### Initialize logger in `main.go`
```go
import "your_project/logger"

func main() {
    logger.InitLogger(true) // or false for production
    logger.Log.Infow("gRPC server started", "port", 50051)
}
```

Replace all `log.Println` or `fmt.Println` with `logger.Log.Infof`, `logger.Log.Errorw`, etc.

---

## 🌍 2. Prometheus Metrics Setup

### Install Prometheus Go client
```bash
go get github.com/prometheus/client_golang/prometheus
```

### Create a metrics package
```go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    RequestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "user_service_requests_total",
            Help: "Total number of requests to user-service, labeled by endpoint",
        },
        []string{"endpoint"},
    )

    RequestDurationHistogram = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "user_service_request_duration_seconds",
            Help:    "Histogram of response durations for user-service requests, labeled by endpoint",
            Buckets: prometheus.DefBuckets,
        },
        []string{"endpoint"},
    )
)

func InitMetrics() {
    prometheus.MustRegister(RequestCounter)
    prometheus.MustRegister(RequestDurationHistogram)
}
```

### Start HTTP server for Prometheus scrape in `main.go`
```go
import (
    "net/http"
    "your_project/metrics"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    metrics.InitMetrics()

    go func() {
        http.Handle("/metrics", promhttp.Handler())
        http.ListenAndServe(":9090", nil)
    }()

    // continue gRPC server setup...
}
```

### Add metrics in gRPC handlers
```go
func (s *Server) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.RegisterUserResponse, error) {
    timer := prometheus.NewTimer(metrics.RequestDurationHistogram.WithLabelValues("RegisterUser"))
    defer timer.ObserveDuration()

    metrics.RequestCounter.WithLabelValues("RegisterUser").Inc()

    // Business logic
}
```

---

## 🚀 3. Deploy Prometheus in Kubernetes

### Add Prometheus Helm repo
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

### Install Prometheus
```bash
helm install prometheus prometheus-community/prometheus
```

### Port-forward Prometheus dashboard
```bash
kubectl port-forward deployment/prometheus-server 9090:9090
```

Access Prometheus UI at `http://localhost:9090`

---

## 📊 4. Expose Microservice Metrics to Prometheus

In `Service.yaml` or Helm chart of user-service:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/path: /metrics
    prometheus.io/port: "9090"
spec:
  selector:
    app: user-service
  ports:
    - protocol: TCP
      port: 9090
      targetPort: 9090
```

Then apply or upgrade:
```bash
kubectl apply -f user-service.yaml
# or
helm upgrade user-service ./charts/user-service
```

---

## 🌌 5. Install and Connect Grafana

### Add Grafana Helm repo
```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
```

### Install Grafana
```bash
helm install grafana grafana/grafana
```

### Get admin password
```bash
kubectl get secret grafana -o jsonpath="{.data.admin-password}" | base64 --decode
```

### Port-forward Grafana
```bash
kubectl port-forward deployment/grafana 3000:3000
```

Access Grafana UI at `http://localhost:3000` (Username: `admin`, Password: retrieved password)

---

## 🔹 6. Connect Grafana to Prometheus

1. Go to Grafana Dashboard > Settings > Data Sources > Add Data Source.
2. Choose Prometheus.
3. Set URL:
```bash
http://prometheus-server.default.svc.cluster.local
```
4. Save & Test.

Now you can build dashboards based on `user_service_requests_total` and other metrics!

---

# 🏅 Congratulations!
You have a full observability stack for your Go microservice: Structured Logging + Metrics + Monitoring!

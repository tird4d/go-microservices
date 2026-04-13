
# � Observability Reference: Prometheus + Grafana + Jaeger

Full picture of how observability works in this project — from a single request all the way to a Grafana dashboard. Use this as a refresh guide.

---

## The Big Picture

```
HTTP/gRPC Request
       ↓
┌─────────────────────────────────────────┐
│  gRPC Interceptor (runs on EVERY call)  │
│   • starts Jaeger span (trace)          │
│   • starts Prometheus timer (latency)   │
│   • increments request counter          │
│   • logs the call (zap)                 │
└─────────────────────────────────────────┘
       ↓
  actual handler (Login, GetProduct...)
       ↓
┌─────────────────────────────────────────┐
│  Interceptor finishes                   │
│   • stops timer → writes to histogram   │
│   • sends span to Jaeger                │
└─────────────────────────────────────────┘
       ↓
  /metrics HTTP endpoint (port 2112/2113)
       ↓
  Prometheus scrapes every 15s
       ↓
  Grafana queries Prometheus with PromQL
       ↓
  Dashboard: Traffic, Latency, Saturation
```

---

## The 3 Files Pattern (add to any new service)

### 1. `metrics/metrics.go` — define what you measure

```go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    // Counts total requests — always goes up, never down
    RequestCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auth_service_requests_total",
            Help: "Total number of requests to auth-service, labeled by endpoint",
        },
        []string{"endpoint"},
    )

    // Records how long each request took, stored in buckets
    // Used to calculate p50, p99 latency
    RequestDurationHistogram = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "auth_service_request_duration_seconds",
            Buckets: prometheus.DefBuckets, // 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1s, 2.5s, 5s, 10s
        },
        []string{"endpoint"},
    )
)

// InitMetrics registers metrics with the default global Prometheus registry.
// Must be called before any metric is used.
func InitMetrics() {
    prometheus.MustRegister(RequestCounter)
    prometheus.MustRegister(RequestDurationHistogram)
}
```

### 2. `interceptors/server.go` — write to metrics on every request

```go
// ============ METRICS ============
// timer automatically records duration when the function returns (via defer)
timer := prometheus.NewTimer(metrics.RequestDurationHistogram.WithLabelValues(info.FullMethod))
defer timer.ObserveDuration()

// increment total request count for this endpoint
metrics.RequestCounter.WithLabelValues(info.FullMethod).Inc()

// ============ EXECUTE HANDLER ============
resp, err := handler(ctx, req)
// timer.ObserveDuration() fires here automatically via defer
```

> ⚠️ **Never put metrics in handlers directly.** The interceptor instruments all endpoints automatically — you write it once and never touch it again.

### 3. `main.go` — start the metrics HTTP server

```go
// InitMetrics must be called before the goroutine starts
metrics.InitMetrics()

// Non-blocking: runs alongside the gRPC server
// Prometheus scrapes this endpoint every 15s
go func() {
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":2112", nil) // each service uses a different port
}()

// Blocking: gRPC server runs on main goroutine
grpcServer.Serve(lis)
```

**Ports used in this project:**
| Service | gRPC port | Metrics port (container) | Host port |
|---|---|---|---|
| user-service | 50051 | 2112 | 2114 |
| auth-service | 50052 | 2112 | 2112 |
| product-service | 50053 | 2112 | 2113 |

> All services use **2112 internally** (standard). Host ports differ only to avoid conflicts when running locally.

---

## Prometheus — the time-series database

**How it works:** Prometheus is a **pull-based** system. It scrapes `/metrics` from each service on a schedule. Services do NOT push data to Prometheus.

**Config file:** `prometheus/prometheus.yml`
```yaml
global:
  scrape_interval: 15s   # scrape every service every 15 seconds

scrape_configs:
  - job_name: 'auth-service'
    static_configs:
      - targets: ['auth-service:2112']   # Docker service name + metrics port

  - job_name: 'product-service'
    static_configs:
      - targets: ['product-service:2113']
```

**Where data is stored:**
- Locally: inside the container at `/prometheus/` (lost on `docker compose down`)
- Production (EKS): PersistentVolumeClaim backed by EBS — survives pod restarts
- Default retention: **15 days**

**Key PromQL queries:**
```promql
# Requests per second (use this, not the raw counter)
rate(auth_service_requests_total[1m])

# p99 latency (99% of requests finished under X seconds)
histogram_quantile(0.99, rate(auth_service_request_duration_seconds_bucket[1m]))

# p50 latency (median)
histogram_quantile(0.50, rate(auth_service_request_duration_seconds_bucket[1m]))

# Average latency
rate(auth_service_request_duration_seconds_sum[5m])
/ rate(auth_service_request_duration_seconds_count[5m])

# Goroutine count (saturation signal)
go_goroutines{job="auth-service"}
```

**Counter vs rate — why it matters:**
- Raw counter: `3, 4, 5, 6` — just goes up, useless for a chart
- `rate()[1m]`: `0.01, 0.04, 0.02` — actual requests/second, shows spikes

---

## Grafana — the visualization layer

**Key fact:** Grafana stores **nothing**. It only queries Prometheus on demand and renders charts. All data lives in Prometheus.

**Where dashboards are stored:**
- `grafana/provisioning/dashboards/microservices.json` — version controlled in Git
- `grafana/provisioning/datasources/prometheus.yml` — auto-connects to Prometheus on startup
- Docker volume `grafana-data` — stores user settings, persists across restarts

**Auto-provisioning (no manual setup needed):**
```yaml
# grafana/provisioning/datasources/prometheus.yml
datasources:
  - name: Prometheus
    type: prometheus
    url: http://prometheus:9090   # Docker service name
    isDefault: true
```

**4 Golden Signals dashboard panels:**

| Panel | Signal | PromQL |
|---|---|---|
| Traffic | req/s per endpoint | `rate(..._total[1m])` |
| Latency p99/p50 | response time | `histogram_quantile(0.99, rate(..._bucket[1m]))` |
| Total Requests | cumulative count | `sum(..._total)` |
| Saturation | goroutine count | `go_goroutines` |

---

## Jaeger — distributed tracing

**Key difference from Prometheus:**
- Prometheus answers: "how many requests? how slow on average?"
- Jaeger answers: "show me exactly what happened during THIS specific request"

**How it works:** Jaeger is a **push-based** system. Services send spans directly to Jaeger via gRPC (OTLP protocol).

**The 3 files for Jaeger:**

`tracing/tracer.go` — connects to Jaeger, sets the global tracer
```go
// Sends spans to jaeger:4317 using OTLP/gRPC
// AlwaysSample = record 100% of requests
// Production: use TraceIDRatioBased(0.1) to sample only 10%
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),   // batches spans, sends every few seconds
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
)
otel.SetTracerProvider(tp)            // set global — interceptor picks this up
```

`interceptors/server.go` — creates a span per request
```go
tracer := otel.Tracer("auth-service")  // gets the global provider
ctx, span := tracer.Start(ctx, info.FullMethod)
defer span.End()
// if error → span.RecordError(err)
```

`interceptors/client.go` — injects trace ID into outgoing calls
```go
// When auth-service calls user-service, this injects the trace ID
// into gRPC headers so Jaeger connects both spans into one trace
otel.GetTextMapPropagator().Inject(ctx, carrier)
```

**No client interceptor needed for Prometheus** — each service measures itself independently. Jaeger needs the client interceptor because a trace must follow the request across multiple services.

---

## Local Docker Compose Setup

Start everything:
```bash
docker compose up -d
```

| Service | URL | Credentials |
|---|---|---|
| Grafana | http://localhost:3001 | admin / admin |
| Prometheus | http://localhost:9090 | none |
| Jaeger UI | http://localhost:16686 | none |
| auth metrics | http://localhost:2112/metrics | none |
| product metrics | http://localhost:2113/metrics | none |

Verify all Prometheus targets are up:
```bash
curl -s "http://localhost:9090/api/v1/targets" | \
  python3 -c "import sys,json; [print(t['labels']['job'], '-', t['health']) for t in json.load(sys.stdin)['data']['activeTargets']]"
```

---

## Kubernetes / EKS Deployment

### Install kube-prometheus-stack (Prometheus + Grafana + AlertManager in one chart)
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace
```

### Expose service metrics with ServiceMonitor CRD
```yaml
# charts/auth-service/templates/servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: auth-service
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: auth-service
  endpoints:
    - port: metrics      # named port in your Service
      path: /metrics
      interval: 15s
```

### Add metrics port to your Helm Service
```yaml
# charts/auth-service/templates/service.yaml
ports:
  - name: grpc
    port: 50052
  - name: metrics       # ServiceMonitor references this name
    port: 2112
```

### Access Grafana on EKS
```bash
kubectl port-forward svc/monitoring-grafana 3000:80 -n monitoring
kubectl get secret monitoring-grafana -n monitoring \
  -o jsonpath="{.data.admin-password}" | base64 --decode
```

---

## Study Resources

| Resource | What it covers | Time |
|---|---|---|
| [PromQL basics](https://prometheus.io/docs/prometheus/latest/querying/basics/) | rate(), histogram_quantile(), labels | 1h |
| [Grafana tutorials](https://grafana.com/tutorials/) | Variables, alerts, annotations | 2h |
| [Google SRE — 4 Golden Signals](https://sre.google/sre-book/monitoring-distributed-systems/) | The theory behind Traffic/Latency/Errors/Saturation | 15min |
| **Prometheus: Up & Running** (Brian Brazil, O'Reilly) | Deep PromQL, label design, histogram internals | book |

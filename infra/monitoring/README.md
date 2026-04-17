# Monitoring Stack — kube-prometheus-stack

This folder documents the **kube-prometheus-stack** which installs Prometheus, Grafana,
AlertManager, and the Prometheus Operator into the cluster in a single Helm install.

> **This is a cluster-level tool, not an app chart.**  
> It is installed once per cluster by hand, not by CI/CD.  
> The upstream Helm chart lives in the `prometheus-community` repo, not in `charts/`.

---

## What gets installed

One `helm install` creates all of this:

```
kube-prometheus-stack (namespace: monitoring)
    ├── Prometheus          — scrapes /metrics from every ServiceMonitor
    ├── Grafana             — dashboards on top of Prometheus data
    ├── AlertManager        — routes alerts to Slack, email, PagerDuty, etc.
    ├── Prometheus Operator — watches for ServiceMonitor CRDs and auto-wires them
    ├── node-exporter       — exports CPU, RAM, disk, network from each EC2 node
    └── kube-state-metrics  — exports Kubernetes object state (pod restarts, etc.)
```

**Why one chart instead of installing each piece separately?**

The Prometheus Operator is the key. It watches the cluster for `ServiceMonitor` objects
(a custom CRD). When your app chart creates a `ServiceMonitor`, Prometheus automatically
starts scraping it — no manual config file edits, no Prometheus restart.

**Comparison to local docker-compose setup:**

| docker-compose | EKS |
|---|---|
| `prometheus.yml` with `static_configs` | `ServiceMonitor` CRD per service |
| Manual `prometheus.yml` edit to add a service | Deploy a `ServiceMonitor`, Operator picks it up |
| `grafana/provisioning/dashboards/*.json` | ConfigMap → Grafana sidecar loads it |
| `alertmanager/alertmanager.yml` | `AlertmanagerConfig` CRD or Helm values |

---

## How scraping works on EKS

```
Your Go service pod
    ↓  exposes :2112/metrics (gRPC interceptor counts requests)
Service (type: ClusterIP)
    ↓  named port "metrics" → 2112
ServiceMonitor  ← your app's Helm chart creates this
    ↓  selector matches the Service labels
Prometheus Operator  ← watches all ServiceMonitors
    ↓  tells Prometheus: "scrape that Service on port 'metrics' every 15s"
Prometheus  ← pulls metrics
    ↓  stores time-series data
Grafana  ← queries Prometheus via PromQL
    ↓
Dashboard panels (Traffic, Latency p99, Goroutines ...)
```

---

## Prerequisites

- `helm` installed locally (`helm version`)
- `kubectl` connected to the EKS cluster (`kubectl config current-context` → `go-microservices`)
- At least **2 nodes** running

```bash
# Check cluster is online and nodes are Ready
kubectl get nodes
# Expected:
# NAME                                         STATUS   ROLES    AGE
# ip-10-0-x-x.eu-central-1.compute.internal   Ready    <none>   5m

# Scale up if needed
eksctl scale nodegroup \
  --cluster=go-microservices \
  --name=ng-1 \
  --nodes=2 \
  --region=eu-central-1
```

---

## Installation

### Step 1 — Add the Helm repo

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

This only needs to be done once per machine. To verify it was added:

```bash
helm repo list
# NAME                    URL
# prometheus-community    https://prometheus-community.github.io/helm-charts
```

### Step 2 — Inspect what the chart will install (optional but educational)

```bash
# See all configurable values before installing
helm show values prometheus-community/kube-prometheus-stack | less

# Interesting ones to look for:
#   grafana.adminPassword
#   grafana.service.type
#   prometheus.prometheusSpec.retention
#   alertmanager.config.receivers
```

### Step 3 — Install the stack

```bash
helm upgrade --install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set grafana.adminPassword=admin \
  --wait --timeout 5m
```

What `--wait` does: blocks until all pods are Running/Ready or the timeout expires.
This takes ~2-3 minutes the first time (image pulls).

Verify everything is running:

```bash
kubectl get pods -n monitoring
# Expected (all should be Running):
# alertmanager-monitoring-kube-prometheus-alertmanager-0   2/2   Running
# monitoring-grafana-<hash>                                3/3   Running
# monitoring-kube-prometheus-operator-<hash>               1/1   Running
# monitoring-kube-prometheus-prometheus-0                  2/2   Running
# monitoring-kube-state-metrics-<hash>                     1/1   Running
# monitoring-prometheus-node-exporter-<hash>               1/1   Running  (one per node)
```

### Step 4 — Access the UIs via port-forward

Unlike local docker-compose (where ports are mapped to localhost), on EKS you use
`kubectl port-forward` to reach cluster services from your machine:

```bash
# Grafana — http://localhost:3001
kubectl port-forward svc/monitoring-grafana 3001:80 -n monitoring

# Prometheus — http://localhost:9090
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090:9090 -n monitoring

# AlertManager — http://localhost:9093
kubectl port-forward svc/monitoring-kube-prometheus-alertmanager 9093:9093 -n monitoring
```

> Run each port-forward in a **separate terminal tab**. They block until you Ctrl+C.

Log in to Grafana at `http://localhost:3001` → user: `admin` / password: `admin`

---

## Wiring your services (ServiceMonitor)

Out of the box, kube-prometheus-stack only scrapes Kubernetes internals (nodes, pods, etc.).
To scrape your Go services, each service needs two small additions to its Helm chart.

### Change 1 — Add a named `metrics` port to the Service

Edit `charts/<service>/templates/service.yaml`. Add the metrics port alongside the existing gRPC port:

```yaml
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: {{ .Values.service.port }}
      protocol: TCP
      name: grpc
    - port: 2112           # ← add this
      targetPort: 2112
      protocol: TCP
      name: metrics         # ← the name "metrics" is referenced by ServiceMonitor below
```

### Change 2 — Create a ServiceMonitor template

Create a new file `charts/<service>/templates/servicemonitor.yaml`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: {{ include "<service>.fullname" . }}
  namespace: monitoring          # must be in the monitoring namespace
  labels:
    release: monitoring          # CRITICAL — must match kube-prometheus-stack's selector
    {{- include "<service>.labels" . | nindent 4 }}
spec:
  namespaceSelector:
    matchNames:
      - prod                     # namespace where your service pods run
  selector:
    matchLabels:
      {{- include "<service>.selectorLabels" . | nindent 6 }}
  endpoints:
    - port: metrics              # matches the named port in service.yaml
      path: /metrics
      interval: 15s
```

> **Why `release: monitoring`?**  
> The Prometheus Operator only watches ServiceMonitors that carry the label
> `release: monitoring` (the Helm release name you used in Step 3).  
> Without this label, Prometheus never discovers your service.

### Verify Prometheus found your service

After deploying your updated chart:

```bash
# Open Prometheus UI
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9090:9090 -n monitoring

# Go to http://localhost:9090 → Status → Targets
# Your service should appear with State = UP
```

Or from the CLI:

```bash
kubectl exec -n monitoring \
  $(kubectl get pod -n monitoring -l app=prometheus -o name | head -1) \
  -- wget -qO- http://localhost:9090/api/v1/targets \
  | python3 -m json.tool | grep '"job"'
# Should include "auth-service", "product-service", "user-service"
```

---

## Import the 4 Golden Signals dashboard

Our custom dashboard JSON lives at `grafana/provisioning/dashboards/microservices.json`.
On EKS we import it manually through the Grafana UI (or via a ConfigMap for automation).

**Manual import (recommended while learning):**

```bash
# 1. Open Grafana
kubectl port-forward svc/monitoring-grafana 3001:80 -n monitoring

# 2. Go to http://localhost:3001 → Dashboards → New → Import

# 3. Click "Upload dashboard JSON file"
#    Select: grafana/provisioning/dashboards/microservices.json

# 4. Set data source → Prometheus → Import
```

**Via ConfigMap (automated, optional):**

```bash
kubectl create configmap microservices-dashboard \
  --from-file=microservices.json=grafana/provisioning/dashboards/microservices.json \
  --namespace monitoring

kubectl label configmap microservices-dashboard \
  grafana_dashboard=1 \
  --namespace monitoring
# Grafana's sidecar picks this up automatically within ~30 seconds
```

---

## Configure AlertManager + Alert Rules

The values file `infra/monitoring/values.yaml` configures both the **alert rules**
(what fires) and **AlertManager** (where notifications go) in one Helm upgrade.

### Step 1 — Apply the values file

```bash
helm upgrade monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --reuse-values \
  -f infra/monitoring/values.yaml
```

> `--reuse-values` keeps all previously set values (like `grafana.adminPassword`)
> and only overrides what is in your file.

### Step 2 — Verify alert rules loaded

```bash
kubectl port-forward svc/monitoring-kube-prometheus-prometheus 9091:9090 -n monitoring
```

Go to `http://localhost:9091` → **Alerts** tab. You should see:

| Alert | Severity | Fires when |
|---|---|---|
| `ServiceDown` | critical | Prometheus can't scrape a service for 1 min |
| `HighLatencyP99` | warning | p99 > 500ms for 2 min on any service |
| `PodCrashLooping` | critical | A pod restarts > 3 times in 15 min |

### Step 3 — Test an alert fires

```bash
# Scale a service to 0 replicas → triggers ServiceDown after 1 minute
kubectl scale deployment user-service --replicas=0 -n prod

# Open AlertManager UI to see it fire
kubectl port-forward svc/monitoring-kube-prometheus-alertmanager 9093:9093 -n monitoring
# → http://localhost:9093

# Restore the service
kubectl scale deployment user-service --replicas=1 -n prod
# Alert resolves within ~1 minute
```

### Step 4 — Enable Slack notifications (optional)

Edit `infra/monitoring/values.yaml` — find the commented Slack block and fill in your webhook:

```yaml
receivers:
  - name: 'slack'
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
        channel: '#alerts'
        send_resolved: true
        title: '{{ .GroupLabels.alertname }} — {{ .CommonLabels.severity }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

Then change the `receiver:` fields in the `route` section from `'null'` to `'slack'`,
and re-run the helm upgrade command from Step 1.

---

# Jaeger — Distributed Tracing

This section documents installing **Jaeger**, a distributed tracing system that collects
traces from all services and visualizes request flows across your microservices.

> **This is also a cluster-level tool.**  
> Installed once per cluster by hand. The Helm chart lives in `jaegertracing` repo.

---

## What is Jaeger?

Jaeger collects **traces** — records of how a request flows through your services.

```
User request
    ↓  (gRPC to API Gateway)
API Gateway traces → span 1 (handles routing)
    ↓  (calls Auth Service)
Auth Service traces → span 2 (validates token)
    ↓  (returns result)
API Gateway spans 3, 4, ... (response handling)
    ↓
Jaeger UI: "This request took 245ms total
           — Auth Service took 50ms
           — Product Service took 120ms
           — Database call took 45ms"
```

**Why it matters:** Find which service is slow, see exact call sequences, debug distributed failures.

---

## Prerequisites

- `helm` installed locally
- `kubectl` connected to EKS cluster
- EKS cluster running (1+ node)
- All services already have OpenTelemetry SDK configured
  (see each service's `config/` for `JAEGER_ENDPOINT` env var)

---

## Installation

### Step 1 — Add the Jaeger Helm repo

```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm repo update
```

### Step 2 — Install Jaeger all-in-one

```bash
helm install jaeger jaegertracing/jaeger \
  --namespace tracing \
  --create-namespace \
  --set allInOne.enabled=true \
  --set provisionDataStore.cassandra=false \
  --set storage.type=memory \
  --wait --timeout 5m
```

**What this installs:**
- `allInOne.enabled=true` — Single pod with Jaeger agent, collector, query UI combined
- `storage.type=memory` — Traces stored in pod RAM (ephemeral, reset on pod restart)
- `cassandra=false` — Don't create a Cassandra DB (for test/dev, lighter weight)

**Why memory storage for now?**
- Sufficient for dev/testing up to ~100K traces before memory pressure
- No database setup needed (fast iteration)
- For production, change to `storage.type: elasticsearch` and provision ES cluster

### Step 3 — Verify Jaeger is running

```bash
kubectl get pods -n tracing
# Expected:
# jaeger-<hash>   1/1   Running

# Check the service is up
kubectl get svc -n tracing
# jaeger        ClusterIP   10.x.x.x    <none>   16686/TCP,6831/UDP,6832/UDP,14268/TCP
```

### Step 4 — Access the Jaeger UI

```bash
# Port-forward to the Jaeger query UI (port 16686)
kubectl port-forward svc/jaeger 16686:16686 -n tracing
```

Then open `http://localhost:16686` in your browser.

You should see a dropdown of services on the left side. Initially it will be empty
until your services send traces.

---

## Wiring services to Jaeger

All 4 services (auth-service, product-service, user-service, api-gateway) are already
configured in their Go code to use OpenTelemetry. The Helm chart values determine
**where** traces go.

### What's already set

Each service's `charts/<service>/values.yaml` has:

```yaml
env:
  jaegerEndpoint: "jaeger-all-in-one.tracing.svc.cluster.local:4317"
```

This tells the service's OpenTelemetry SDK to send traces to the Jaeger collector
running in the `tracing` namespace at port 4317 (gRPC OTLP).

### Deploy a service and watch traces appear

```bash
# If you haven't deployed services yet:
helm install auth-service charts/auth-service \
  --namespace prod \
  --create-namespace \
  -f charts/auth-service/values.yaml

# Send a request to generate traces
kubectl port-forward svc/api-gateway 8000:8000 -n prod
# In another terminal:
curl -X POST http://localhost:8000/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'

# Open Jaeger UI
kubectl port-forward svc/jaeger 16686:16686 -n tracing
# → http://localhost:16686
# Select service from dropdown (e.g., "api-gateway")
# Traces should appear within a few seconds
```

---

## Jaeger Configuration

### Using Elasticsearch for persistence (optional)

For production, store traces in Elasticsearch instead of RAM:

```bash
# First, install Elasticsearch (or use Amazon OpenSearch)
helm install elasticsearch elastic/elasticsearch \
  --namespace tracing \
  --set replicas=1 \
  --set minimumMasterNodes=1

# Then, upgrade Jaeger to use it
helm upgrade jaeger jaegertracing/jaeger \
  --namespace tracing \
  --reuse-values \
  --set storage.type=elasticsearch \
  --set storage.elasticsearch.host=elasticsearch \
  --set storage.elasticsearch.port=9200 \
  --set storage.elasticsearch.scheme=http
```

### Memory limits (important for in-memory storage)

If using `storage.type=memory`, monitor Jaeger memory:

```bash
kubectl top pod -n tracing
# NAME                MEMORY
# jaeger-<hash>       256Mi (check if approaching limit)

# If needed, increase:
helm upgrade jaeger jaegertracing/jaeger \
  --namespace tracing \
  --reuse-values \
  --set allInOne.resources.limits.memory=1Gi
```

### Sampling (reduce trace volume)

By default, Jaeger samples 1 out of every 100 traces. Adjust in values:

```bash
helm upgrade jaeger jaegertracing/jaeger \
  --namespace tracing \
  --reuse-values \
  --set allInOne.samplingConfig='{"default_strategy":{"type":"const","param":1}}'
# type: "const", param: 1 = sample 100% (for dev)
# type: "probabilistic", param: 0.1 = sample 10%
```

---

## Viewing traces

### Common queries in Jaeger UI

1. **Find slow requests:**
   - Select service → Click "Trace duration" duration filter
   - Set min/max: e.g., `>500ms`

2. **Find errors:**
   - Select service → Filter by tags: `error=true`

3. **Compare two traces:**
   - Click on one trace, then Ctrl+Click another
   - View side-by-side comparison

4. **Follow a user's requests:**
   - Click a trace tag → filter by `user_id=123`

---

## Uninstalling Jaeger

```bash
# Remove the Jaeger release
helm uninstall jaeger -n tracing

# Remove namespace and all data
kubectl delete namespace tracing
```

---

## Uninstallation (Full Monitoring Stack)

```bash
# Remove both Jaeger and kube-prometheus-stack
helm uninstall jaeger -n tracing
helm uninstall monitoring -n monitoring

# Remove the namespaces
kubectl delete namespace tracing monitoring
```

> No AWS cost impact — unlike ingress-nginx, kube-prometheus-stack does **not** create
> an AWS LoadBalancer. All services are ClusterIP (accessed via port-forward).

---

## Upgrading

```bash
# Pull latest chart metadata
helm repo update

# See what version is available
helm search repo prometheus-community/kube-prometheus-stack

# Upgrade in-place (keeps your data)
helm upgrade monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --reuse-values
```

---

## Current state

| Item | Value |
|---|---|
| Helm release name | `monitoring` |
| Namespace | `monitoring` |
| Chart version | install to record |
| Grafana admin password | `admin` (change on first install) |
| App metrics port | `:2112` (all services) |
| Local docker-compose Grafana | `localhost:3001` |
| Local docker-compose Prometheus | `localhost:9090` |
| Local docker-compose AlertManager | `localhost:9093` |
| EKS (via port-forward) Grafana | `localhost:3001` |
| EKS (via port-forward) Prometheus | `localhost:9090` |

---

## Troubleshooting

**Pods stuck in `Pending`:**
```bash
kubectl describe pod <pod-name> -n monitoring
# Look for: "Insufficient memory" or "no nodes available"
# Fix: scale up nodegroup to 2+ nodes
```

**ServiceMonitor not picked up (target missing in Prometheus):**
```bash
# Check the label selector the Operator is using
kubectl get prometheus -n monitoring -o yaml | grep serviceMonitorSelector -A5
# Output will show: matchLabels: release: monitoring
# Make sure your ServiceMonitor has that label
```

**Grafana shows "No data" on dashboard panels:**
```bash
# 1. Confirm Prometheus target is UP
#    http://localhost:9090 → Status → Targets

# 2. Check the metric name in the panel query matches what your service exports
#    auth_service_requests_total  (auth service)
#    product_service_requests_total  (product service)

# 3. Confirm time range in Grafana is correct (top-right corner)
#    Set to "Last 15 minutes" to start
```

**AlertManager not receiving alerts:**
```bash
kubectl logs -n monitoring \
  $(kubectl get pod -n monitoring -l app.kubernetes.io/name=alertmanager -o name | head -1)
# Look for connection errors or config parse failures
```

**Port-forward drops after a few minutes:**
```bash
# Normal on EKS — just re-run the port-forward command
# For persistent access, consider a LoadBalancer service or ingress rule for Grafana
```

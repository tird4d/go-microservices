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

## Configure AlertManager (optional)

The stack ships with a default AlertManager config. To add Slack notifications,
override the config via Helm values:

Create a file `infra/monitoring/values.yaml`:

```yaml
alertmanager:
  config:
    global:
      resolve_timeout: 5m
    route:
      group_by: ['job', 'severity']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 12h
      receiver: 'slack'
    receivers:
      - name: 'slack'
        slack_configs:
          - api_url: 'https://hooks.slack.com/services/YOUR/WEBHOOK/URL'
            channel: '#alerts'
            title: '{{ .GroupLabels.job }} — {{ .CommonLabels.severity }}'
            text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
    inhibit_rules:
      - source_matchers:
          - severity="critical"
        target_matchers:
          - severity="warning"
        equal: ['job']
```

Then upgrade the release with the values file:

```bash
helm upgrade monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --reuse-values \
  -f infra/monitoring/values.yaml
```

> `--reuse-values` keeps all previously set values (like `grafana.adminPassword`)
> and only overrides what is in your file.

---

## Uninstallation

```bash
# Remove the stack
helm uninstall monitoring -n monitoring

# Remove the namespace (deletes all PVCs and stored data too)
kubectl delete namespace monitoring
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

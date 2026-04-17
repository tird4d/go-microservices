# EKS Cluster

This cluster is created using `eksctl` on top of an existing VPC provisioned via Terraform (`infra/vpc`).

**Cluster Config:**
- Kubernetes version: 1.35
- Node type: t3.small (desiredCapacity: 1, min: 1, max: 2)
- Private networking enabled
- Auto-scaling ready via Cluster Autoscaler

---

## Core Infrastructure Components

| Component | Purpose | Namespace |
|---|---|---|
| **ingress-nginx** | Ingress controller (LoadBalancer) | `ingress-nginx` |
| **metrics-server** | Kubelet metrics for HPA/Cluster Autoscaler | `kube-system` |
| **VPC CNI** | AWS networking addon | `kube-system` |

---

## Auto-Scaling Features

### 1. Horizontal Pod Autoscaler (HPA)

#### What is HPA?

HPA automatically scales the number of pod replicas (horizontal scaling) based on observed metrics like CPU or memory usage. Instead of manually adding replicas, Kubernetes watches metrics and adjusts replica count automatically.

**Benefits:**
- ✅ Handles traffic spikes without manual intervention
- ✅ Reduces cost by scaling down during low usage
- ✅ Maintains performance targets (SLA)
- ✅ Works seamlessly with Cluster Autoscaler

#### How HPA Works

```
1. Metrics Server collects metrics from all pods/nodes every ~15s
2. HPA Controller queries Metrics Server
3. HPA compares current metric vs target (e.g., current CPU 80% vs target 70%)
4. If metric > target: Scale UP (add replicas, up to maxReplicas)
5. If metric < target: Scale DOWN (remove replicas, down to minReplicas)
6. Wait cooldown period (~3min for scale-down, ~3min for scale-up)
7. Repeat
```

#### HPA Configuration in This Cluster

All microservices have HPA enabled via Helm charts (`charts/*/values.yaml`):

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

**What this means:**
- Services start with **1 replica** per service
- HPA monitors CPU usage
- When CPU **> 70%**: Add replicas (up to 10 max)
- When CPU **< 70%** for 5 minutes: Remove replicas (down to 1 min)

#### Installation & Prerequisites

HPA requires **Metrics Server** to collect pod metrics. This was installed during cluster setup:

```bash
# Verify Metrics Server is running
kubectl get deployment metrics-server -n kube-system

# Test metrics collection (wait ~1 min after pod start)
kubectl top nodes
kubectl top pods -n prod
```

If metrics aren't available (showing `<unknown>`), Metrics Server may not be running:

```bash
# Reinstall if needed
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm install metrics-server metrics-server/metrics-server \
  --namespace kube-system \
  --set args={--kubelet-insecure-tls}
```

#### Verify HPA Status

**Current HPA objects:**
```bash
kubectl get hpa -n prod
```

**Output example:**
```
NAME              REFERENCE                    TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
api-gateway       Deployment/api-gateway       45%/70%   1         10        1          2h
auth-service      Deployment/auth-service      30%/70%   1         10        1          2h
product-service   Deployment/product-service   52%/70%   1         10        2          2h
user-service      Deployment/user-service      25%/70%   1         10        1          2h
```

**Detailed HPA info:**
```bash
kubectl describe hpa api-gateway -n prod
```

**Watch HPA activity in real-time:**
```bash
# See scaling events as they happen
kubectl get hpa -n prod -w

# Example output (columns update as CPU changes):
# NAME         REFERENCE               TARGETS   MINPODS   MAXPODS   REPLICAS   AGE
# api-gateway  Deployment/api-gateway  45%/70%   1         10        2          15m
```

#### Testing HPA Scaling

**Generate load to trigger scale-up:**

```bash
# Get API Gateway pod
GATEWAY_POD=$(kubectl get pods -n prod -l app=api-gateway -o jsonpath='{.items[0].metadata.name}')

# Generate CPU load (run for ~5 minutes)
kubectl exec -it $GATEWAY_POD -n prod -- \
  /bin/sh -c "for i in {1..10}; do (md5sum /dev/zero &) done"
```

**In another terminal, watch HPA:**
```bash
kubectl get hpa -n prod -w
# Watch TARGETS column increase from 45%/70% → 85%/70% → 120%/70%
# Watch REPLICAS column increase from 1 → 2 → 3 as CPU exceeds 70%
```

**Clean up load test:**
```bash
# Kill all background processes in the pod
kubectl exec -it $GATEWAY_POD -n prod -- /bin/sh -c "pkill md5sum || true"
```

#### Scaling Behavior Details

**Scale-Up Behavior:**
- Checks metrics every 15 seconds
- Scales up when metric > target (70%)
- Scale-up cooldown: 3 minutes (scales aggressively under load)
- Can add multiple replicas at once if far above target

**Scale-Down Behavior:**
- Waits 5 minutes of stable low usage before scaling down
- More conservative to avoid "flapping" (constant up/down)
- Only removes 1 replica at a time (gracefully)

**Example timeline:**
```
Time 0:00  → CPU 45% (1 replica) - stable
Time 0:30  → CPU 80% (exceeds 70%) - HPA scale-up triggered
Time 0:45  → Scale-up executed, now 2 replicas
Time 1:15  → CPU 55% (dropped) - starting scale-down timer
Time 6:15  → CPU still 55% for 5 minutes - HPA scale-down triggered
Time 6:20  → Scale-down executed, back to 1 replica
```

#### Advanced HPA Configuration

Custom metrics (beyond CPU/memory):

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  # Optional: Add custom metrics via Prometheus
  # targetMemoryUtilizationPercentage: 80
  # Or custom: targetValue from app metrics
```

For learning, the current CPU-based HPA is sufficient. Custom metrics require Prometheus adapter integration.

#### Troubleshooting HPA

**Problem: HPA shows `<unknown>` for targets**
```bash
# Solution: Wait 1-2 minutes for Metrics Server to collect data
# Check Metrics Server logs
kubectl logs -n kube-system -l k8s-app=metrics-server
```

**Problem: Pods not scaling despite high CPU**
```bash
# Check HPA status
kubectl describe hpa api-gateway -n prod

# Common issues:
# 1. Pod CPU requests/limits not set (HPA uses these for % calculation)
# 2. Metrics Server not running (kubectl get deploy metrics-server -n kube-system)
# 3. Already at maxReplicas (can't scale further)
```

**Problem: Pods rapidly scaling up and down (flapping)**
```bash
# Increase scale-down stabilization window
# Edit HPA and add:
# behavior:
#   scaleDown:
#     stabilizationWindowSeconds: 300  # 5 minutes
```

#### Related Concepts

- **Cluster Autoscaler**: Scales EC2 nodes when pods can't fit (see below)
- **VPA (Vertical Pod Autoscaler)**: Adjusts CPU/memory requests (not used here)
- **Custom Metrics**: Scale based on app-specific metrics (requires Prometheus adapter)

### 2. Cluster Autoscaler

#### What is Cluster Autoscaler?

Cluster Autoscaler automatically scales the **number of EC2 nodes** (infrastructure-level scaling) when pods don't have enough space to run. Unlike HPA which scales pod replicas, Cluster Autoscaler scales the underlying compute capacity.

**When to use:**
- HPA scales pod replicas, but no nodes have capacity → Cluster Autoscaler adds nodes
- Node resources exhausted (CPU/memory reserved) → Cluster Autoscaler adds nodes

**Benefits:**
- ✅ Automatic infrastructure scaling without manual EC2 management
- ✅ Cost optimization: Removes unused nodes
- ✅ Works alongside HPA for full vertical & horizontal scaling
- ✅ AWS Auto Scaling Group (ASG) integration

#### How Cluster Autoscaler Works

```
1. Pod created but no node has capacity (Pending state)
2. Cluster Autoscaler detects Pending pod
3. Calculates required nodes from pod requests
4. Scales ASG from current → needed size (up to max: 2)
5. AWS launches new EC2 instances
6. Cluster Autoscaler reschedules pending pods onto new nodes
7. Monitors node utilization over time (10 minutes)
8. If node usage < threshold: Removes underutilized nodes
```

#### Cluster Autoscaler Setup in This Environment

**Current status:**
- ✅ IRSA (IAM Roles for Service Accounts) configured on cluster
- ✅ ServiceAccount `cluster-autoscaler` created with proper IAM policy
- ✅ EC2 Auto Scaling Group configured: min=1, max=2 nodes
- ⏳ Helm chart installation (next step)

**Why IRSA + Cluster Autoscaler:**
- Pod assumes IAM role via ServiceAccount
- Calls AWS API to scale ASG (no stored credentials)
- Temporary STS credentials auto-rotated

#### IRSA Setup (Already Done — Recorded for Reference)

**Why do we need IRSA for Cluster Autoscaler?**

Cluster Autoscaler runs as a pod inside Kubernetes, but it needs to call AWS APIs to add/remove EC2 nodes (via the Auto Scaling Group). Without credentials, it can't do anything.

The naive solution is to hard-code AWS keys as environment variables in the pod — but that's a security risk (keys can leak via logs, `kubectl describe`, or a compromised pod).

**IRSA (IAM Roles for Service Accounts)** solves this properly:
1. The cluster has an OIDC provider (a trust bridge between K8s and AWS IAM)
2. The Cluster Autoscaler pod uses a Kubernetes ServiceAccount
3. That ServiceAccount is annotated with an IAM Role ARN
4. When the pod starts, AWS automatically injects short-lived STS credentials via a projected volume
5. The pod assumes the IAM role without any stored secrets — credentials rotate automatically every hour

This is the AWS-recommended production pattern. No keys stored anywhere.

---

This was created once with `eksctl`. **Do not run again** — the ServiceAccount and IAM role already exist.

```bash
# Step 1: Create OIDC provider for the cluster (enables IRSA)
# This registers the cluster's internal token issuer as a trusted identity provider in AWS IAM.
# Without this, AWS IAM has no way to verify that a token from this cluster is legitimate.
eksctl utils associate-iam-oidc-provider \
  --cluster=go-microservices \
  --region=eu-central-1 \
  --approve

# Step 2: Create IAM ServiceAccount with AutoScalingFullAccess
# This does two things atomically:
#   a) Creates an IAM Role with a trust policy scoped to this specific ServiceAccount
#      (only pods using cluster-autoscaler SA in kube-system can assume this role)
#   b) Creates a Kubernetes ServiceAccount annotated with that role's ARN
eksctl create iamserviceaccount \
  --cluster=go-microservices \
  --namespace=kube-system \
  --name=cluster-autoscaler \
  --attach-policy-arn=arn:aws:iam::aws:policy/AutoScalingFullAccess \
  --override-existing-serviceaccounts \
  --region=eu-central-1 \
  --approve
```

**What this created:**
- IAM Role: `eksctl-go-microservices-addon-iamserviceaccou-Role1-1vnliWKy9Uk0`
  - Attached policy: `AutoScalingFullAccess`
  - Trust policy: allows the cluster's OIDC provider to assume this role
- K8s ServiceAccount: `cluster-autoscaler` in `kube-system`
  - Annotation: `eks.amazonaws.com/role-arn: arn:aws:iam::114851843413:role/eksctl-go-microservices-addon-iamserviceaccou-Role1-1vnliWKy9Uk0`

**Verify it's still in place:**
```bash
# Check ServiceAccount annotation
kubectl get sa cluster-autoscaler -n kube-system -o jsonpath='{.metadata.annotations}'

# Check IAM role policy
aws iam list-attached-role-policies \
  --role-name eksctl-go-microservices-addon-iamserviceaccou-Role1-1vnliWKy9Uk0
```

#### Install Cluster Autoscaler

```bash
# Add Helm repository
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm repo update

# Install Cluster Autoscaler (IRSA already set up — use existing ServiceAccount)
helm install cluster-autoscaler autoscaler/cluster-autoscaler \
  --namespace kube-system \
  --set autoDiscovery.clusterName=go-microservices \
  --set awsRegion=eu-central-1 \
  --set rbac.serviceAccount.create=false \
  --set rbac.serviceAccount.name=cluster-autoscaler \
  --set extraArgs.balance-similar-node-groups=true \
  --set extraArgs.skip-nodes-with-system-pods=false
```

> **Why `rbac.serviceAccount.create=false`?**
> By default, Helm would create a brand-new ServiceAccount for Cluster Autoscaler. But a new SA would have no IRSA annotation — meaning the pod would start with no AWS credentials and fail silently. We set `create=false` + `name=cluster-autoscaler` to tell Helm: "don't create a SA, just use the existing one that `eksctl` already set up with the IAM role annotation." This is critical — getting this wrong means Cluster Autoscaler runs but can never call AWS.

#### Verify Cluster Autoscaler

```bash
# Check pod is running
kubectl get pods -n kube-system -l app.kubernetes.io/name=cluster-autoscaler

# View logs
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler --tail=50

# Check it detected ASG
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler | grep "Auto Scaling Group"
```

#### Testing Cluster Autoscaler

**Scenario: Create many pods to trigger node scaling**

```bash
# Create deployment with many replicas to exceed node capacity
kubectl create deployment load-test --image=nginx:latest \
  --replicas=30 \
  -n prod

# Watch Cluster Autoscaler in action
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler -f

# Expected flow:
# 1. Pods created, many stay Pending (no space)
# 2. Cluster Autoscaler detects Pending pods
# 3. "Expanding Node Group" message in logs
# 4. Second node starts launching
# 5. Pending pods get scheduled on new node
```

**Watch cluster nodes:**
```bash
# Terminal 1: Watch node count
kubectl get nodes -w

# Terminal 2: Watch pod status
kubectl get pods -n prod -o wide -w
```

**Clean up test:**
```bash
kubectl delete deployment load-test -n prod
# Cluster Autoscaler will eventually remove the unused node (after ~10 min idle)
```

#### HPA + Cluster Autoscaler Interaction

**Full scaling workflow:**
```
1. Traffic increases
2. HPA detects high CPU → Scales pods 1 → 10 replicas
3. Node 1 becomes full (CPU/memory reserved for 10 pods)
4. New pods can't fit (Pending)
5. Cluster Autoscaler detects Pending pods
6. Scales ASG 1 → 2 nodes
7. Cluster Autoscaler reschedules pending pods onto node 2
8. Services now running on both nodes under HPA control
```

**Example timeline:**
```
Time 0:00   → 1 pod per service, 1 node (40% utilized)
Time 1:00   → Load increases, CPU jumps to 85%
Time 1:15   → HPA scales: api-gateway 1→3 replicas
Time 1:30   → CPU drops to 65% (HPA scale satisfied for now)
Time 2:00   → Load stays high, CPU back to 90%
Time 2:15   → HPA scales again: api-gateway 3→6, auth-service 1→4
Time 2:30   → Pod resources exhausted on node 1 (many Pending pods)
Time 2:45   → Cluster Autoscaler adds node 2
Time 3:00   → Pending pods scheduled, now balanced across 2 nodes
Time 30:00  → Load drops, pods scale down, node 2 idle for 10 min
Time 40:00  → Cluster Autoscaler removes node 2 (not needed)
```

#### Monitoring Cluster Autoscaler

**Key metrics to watch:**
```bash
# Node count
kubectl get nodes

# Node capacity
kubectl top nodes

# Pod distribution
kubectl get pods -n prod -o wide --sort-by=.spec.nodeName

# ASG status via AWS CLI
aws autoscaling describe-auto-scaling-groups \
  --auto-scaling-group-names go-microservices-ng-1 \
  --region eu-central-1
```

#### Troubleshooting Cluster Autoscaler

**Problem: Cluster Autoscaler not scaling despite Pending pods**
```bash
# Check logs for errors
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler | grep -i error

# Verify ASG configuration
aws autoscaling describe-auto-scaling-groups \
  --auto-scaling-group-names go-microservices-ng-1 \
  --region eu-central-1

# Common issues:
# 1. ASG max capacity already reached (currently max=2)
# 2. IAM role missing permissions
# 3. Pod requests so large they can't fit even on new nodes
```

**Problem: Nodes not scaling down**
```bash
# Cluster Autoscaler is conservative (waits ~10 min)
# Check unneeded nodes:
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler | grep "Unneeded"

# Force faster scale-down (development only):
# Edit Cluster Autoscaler deployment and add:
# - --scale-down-delay-after-add=5m
# - --scale-down-unneeded-time=5m
```

#### Cluster Autoscaler vs Manual Scaling

**Manual scaling (not recommended):**
```bash
eksctl scale nodegroup \
  --cluster=go-microservices \
  --name=ng-1 \
  --nodes=2 \
  --region=eu-central-1
```

**Benefits of Cluster Autoscaler over manual:**
- ✅ Automatic response to workload changes
- ✅ Cost optimization (removes unused nodes)
- ✅ No SLA violations from capacity issues
- ✅ Production-ready approach

---

#### HPA + Cluster Autoscaler: Complete Scaling Architecture

This cluster implements **two-level auto-scaling:**

| Layer | Component | Scales | Trigger | Min/Max |
|-------|-----------|--------|---------|---------|
| **Pods** | HPA | Pod replicas per service | CPU > 70% | 1-10 replicas per service |
| **Infrastructure** | Cluster Autoscaler | EC2 nodes | Pods can't fit | 1-2 nodes |

**Summary:**
- Pod pressure (HPA) → Cluster Autoscaler adds nodes
- Low pressure → Cluster Autoscaler removes nodes
- Full auto-scaling for dev/learning without manual intervention

---

## Security & Access

### IRSA (IAM Roles for Service Accounts)

All service accounts can assume IAM roles securely without storing credentials:

```bash
# Check IRSA provider is enabled
eksctl utils describe-addon-versions --kubernetes-version=1.35 --region=eu-central-1

# ServiceAccount → IAM Role mapping:
# cluster-autoscaler@kube-system → AutoScalingFullAccess policy
```

**Why IRSA matters:**
- ✅ No hard-coded AWS credentials in pods
- ✅ Automatic credential rotation (AWS STS)
- ✅ Fine-grained IAM permissions per service
- ✅ Production-grade security

---

## Common Operations

### Scale Nodegroup Manually

```bash
eksctl scale nodegroup \
  --cluster=go-microservices \
  --name=ng-1 \
  --nodes=3 \
  --region=eu-central-1
```

### View Cluster Events

```bash
kubectl get events -A --sort-by='.lastTimestamp' | tail -20
```

### Check HPA Activity

```bash
# Watch HPA in real-time
kubectl get hpa -n prod -w

# See detailed HPA status
kubectl describe hpa api-gateway -n prod
```

### Check Node Capacity

```bash
kubectl top nodes
kubectl describe nodes
```

---

## Create

```bash
eksctl create cluster -f cluster.yaml
```

## Delete

```bash
eksctl delete cluster -f cluster.yaml
```

---

## See Also

- **Monitoring & Tracing**: See [`infra/monitoring/README.md`](../monitoring/README.md) for Prometheus, Grafana, and Jaeger setup
- **VPC & Networking**: See [`infra/vpc/README.md`](../vpc/README.md) for network architecture
- **Services**: See [`charts/`](../../charts/) for microservice deployments

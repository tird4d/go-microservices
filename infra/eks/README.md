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

All microservices have HPA enabled:
- **Target metric**: CPU utilization
- **Target threshold**: 70%
- **Min replicas**: 1
- **Max replicas**: 10

**Verify HPA status:**
```bash
kubectl get hpa -n prod
# Shows current CPU usage vs 70% target
```

### 2. Cluster Autoscaler

Automatically scales EC2 nodes (infrastructure) when pods need more capacity.

**Enabled via:**
- ✅ IRSA (IAM Roles for Service Accounts) configured
- ✅ ServiceAccount `cluster-autoscaler` created with IAM policy `AutoScalingFullAccess`
- ✅ Ready for Helm installation

**How it works:**
```
Pod can't fit on existing nodes (Pending)
  → Cluster Autoscaler detects this
  → Scales ASG from 1 → 2 nodes (up to max: 2)
  → Pod gets scheduled
```

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

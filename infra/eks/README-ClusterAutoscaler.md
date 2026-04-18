# Cluster Autoscaler

Cluster Autoscaler automatically scales the **number of EC2 nodes** in the cluster when pods cannot be scheduled due to insufficient resources. It works alongside HPA — HPA scales pod replicas, Cluster Autoscaler scales the underlying EC2 nodes.

---

## How It Works

```
1. HPA adds more replicas (pods) under load
2. No node has enough CPU/memory to schedule the new pods
3. Pods stay in Pending state
4. Cluster Autoscaler detects Pending pods
5. Calls AWS Auto Scaling Group (ASG) API to add a new EC2 node
6. New node joins the cluster
7. Pending pods get scheduled on the new node and become Running
8. After load drops, Cluster Autoscaler removes underutilized nodes (after ~10 min idle)
```

**Without Cluster Autoscaler:** Pending pods just sit forever — no node ever comes.  
**With Cluster Autoscaler:** New EC2 node provisions in ~2 minutes, pods recover automatically.

---

## Live Proof — What We Observed

During load testing with k6 (`auth_service_load_test.js`), HPA scaled auth-service from 1 → 8 replicas.  
The cluster only had capacity for 1 pod. The result:

```
NAME                                READY   STATUS    RESTARTS   AGE
auth-service-69865b6dbf-n8nzn       1/1     Running   0          90m   ← only one fit
auth-service-69865b6dbf-cl94b       0/1     Pending   0          6m49s ← no capacity
auth-service-69865b6dbf-r2qrj       0/1     Pending   0          6m49s
auth-service-69865b6dbf-x8x7t       0/1     Pending   0          6m49s
auth-service-69865b6dbf-45mnh       0/1     Pending   0          6m34s
auth-service-69865b6dbf-4gm2b       0/1     Pending   0          6m34s
auth-service-69865b6dbf-5dvjv       0/1     Pending   0          6m34s
auth-service-69865b6dbf-lhbs8       0/1     Pending   0          6m34s
```

This is exactly the scenario Cluster Autoscaler is designed to fix.

---

## Cluster Configuration

- **Node type:** t3.small
- **ASG min:** 1 node
- **ASG max:** 3 nodes
- **IRSA:** ✅ Configured (cluster-autoscaler ServiceAccount in kube-system)
- **Helm release:** `cluster-autoscaler v9.56.0` (app: v1.35.0) ✅ Installed
- **Verified:** HPA → 8 replicas → Pending pods → CA added node 3 → 8/8 Running ✅

---

## IRSA Setup (Already Done — Reference Only)

Cluster Autoscaler runs as a pod but must call AWS APIs to scale the ASG. IRSA (IAM Roles for Service Accounts) provides it with short-lived, auto-rotating credentials — no stored keys.

> **Do not run these again.** The ServiceAccount and IAM role already exist.

```bash
# Step 1: Associate OIDC provider (enables IRSA on the cluster)
eksctl utils associate-iam-oidc-provider \
  --cluster=go-microservices \
  --region=eu-central-1 \
  --approve

# Step 2: Create IAM ServiceAccount with AutoScaling permissions
eksctl create iamserviceaccount \
  --cluster=go-microservices \
  --namespace=kube-system \
  --name=cluster-autoscaler \
  --attach-policy-arn=arn:aws:iam::aws:policy/AutoScalingFullAccess \
  --override-existing-serviceaccounts \
  --region=eu-central-1 \
  --approve
```

**What was created:**
- IAM Role: `eksctl-go-microservices-addon-iamserviceaccou-Role1-1vnliWKy9Uk0`
- K8s ServiceAccount: `cluster-autoscaler` in `kube-system` (annotated with the IAM role ARN)

**Verify it's still in place:**
```bash
kubectl get sa cluster-autoscaler -n kube-system -o jsonpath='{.metadata.annotations}'
```

---

## Installation

```bash
helm repo add autoscaler https://kubernetes.github.io/autoscaler
helm repo update

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
> Helm would otherwise create a new ServiceAccount with no IRSA annotation, meaning the pod would start but silently fail to call AWS. We reuse the existing SA that `eksctl` set up with the IAM role annotation.

---

## Verify It's Running

```bash
# Pod is running
kubectl get pods -n kube-system -l app.kubernetes.io/name=cluster-autoscaler

# Logs — confirm it detected the ASG
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler | grep -i "auto scaling group\|node group"
```

---

## Testing

**Trigger node scale-up by overloading with replicas:**

```bash
# Create a deployment that exceeds current node capacity
kubectl create deployment ca-test --image=nginx:latest --replicas=30 -n prod

# Terminal 1: Watch nodes
kubectl get nodes -w

# Terminal 2: Watch Cluster Autoscaler logs
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler -f

# Expected:
# 1. Many pods go Pending
# 2. Cluster Autoscaler logs: "Expanding Node Group"
# 3. New EC2 node joins cluster (~2 min)
# 4. Pending pods become Running

# Clean up
kubectl delete deployment ca-test -n prod
# Node scales back down after ~10 min of low utilization
```

**Describe a Pending pod to confirm the cause:**

```bash
kubectl describe pod <pending-pod-name> -n prod | grep -A5 "Events:"
# Shows: "0/N nodes are available: insufficient cpu"
```

---

## HPA + Cluster Autoscaler Together

```
Traffic spike
    │
    ▼
HPA detects CPU > 70%
    │
    ▼
HPA scales pods: 1 → 8 replicas
    │
    ▼
No node has capacity → pods Pending
    │
    ▼
Cluster Autoscaler detects Pending pods
    │
    ▼
Adds new EC2 node (ASG: 1 → 2)
    │
    ▼
Pods scheduled on new node → 8/8 Running
    │
    ▼
Load drops → CPU < 70% for 5 min
    │
    ├─▶ HPA scales pods back: 8 → 1
    │
    └─▶ Node idle for 10 min → Cluster Autoscaler removes node (ASG: 2 → 1)
```

---

## Scale-Down Behavior

| Event | Delay | Reason |
|-------|-------|--------|
| HPA scale-down | 5 minutes | Stabilization window (prevents flapping) |
| Node scale-down | ~10 minutes | Cluster Autoscaler waits to confirm node is truly idle |

---

## Troubleshooting

**Pods still Pending after Cluster Autoscaler is installed:**
```bash
# Check Cluster Autoscaler logs
kubectl logs -n kube-system -l app.kubernetes.io/name=cluster-autoscaler | tail -30

# Common causes:
# 1. ASG max already reached (max: 2 nodes)
# 2. IRSA not working — check "failed to describe" errors in logs
# 3. ASG tags missing — Cluster Autoscaler uses tags to discover ASGs:
#    k8s.io/cluster-autoscaler/go-microservices = owned
#    k8s.io/cluster-autoscaler/enabled = true
```

**Verify ASG tags are set:**
```bash
AWS_PROFILE=home AWS_PAGER="" aws autoscaling describe-auto-scaling-groups \
  --region eu-central-1 \
  --query 'AutoScalingGroups[].{Name:AutoScalingGroupName,Tags:Tags[?starts_with(Key,`k8s.io`)]}'
```

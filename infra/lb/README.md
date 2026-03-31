# Load Balancer — nginx Ingress Controller

This folder documents the **nginx Ingress Controller** which is the only component
that creates a real AWS Load Balancer in this project.

> **This is a cluster-level tool, not an app chart.**  
> It is installed once per cluster by hand, not by CI/CD.  
> The Helm chart for it lives in the upstream `ingress-nginx` repo, not in `charts/`.

---

## How it works

```
Internet
    ↓  HTTP/HTTPS
AWS Classic LoadBalancer   ← created automatically by Kubernetes Cloud Controller
    ↓
nginx Ingress Controller pod  (namespace: ingress-nginx)
    ↓  reads Ingress objects from all namespaces
    ├── /api/*  →  api-gateway:8080       (rule from charts/api-gateway)
    └── /*      →  frontend-service:80    (rule from charts/frontend-service)
```

The AWS LoadBalancer is **not created by Terraform or eksctl**.  
It is created by the **AWS Cloud Controller Manager** (a pod inside the cluster) the moment
Kubernetes sees a `Service` with `type: LoadBalancer` — which the ingress-nginx Helm chart creates.

**Cost:** ~$20/month while running. Always delete it before shutting down the cluster.

---

## Prerequisites

- `helm` installed locally
- `kubectl` connected to the EKS cluster (`kubectl config current-context` → `go-microservices`)
- At least **2 nodes** running (t3.small has a 11-pod limit; ingress controller needs 2 extra pods)

```bash
# Check node count
kubectl get nodes

# Scale up if only 1 node
eksctl scale nodegroup \
  --cluster=go-microservices \
  --name=ng-1 \
  --nodes=2 \
  --region=eu-central-1
```

---

## Installation

```bash
# 1. Add the Helm repo (only needed once on your machine)
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# 2. Install the controller
helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --wait --timeout 3m
```

Wait ~60 seconds for AWS to provision the LoadBalancer, then get the public hostname:

```bash
kubectl get svc -n ingress-nginx
# EXTERNAL-IP column → abc123.eu-central-1.elb.amazonaws.com
```

Verify it is reachable:

```bash
curl -s -o /dev/null -w "%{http_code}" http://<EXTERNAL-IP>/
# Should return 404 (nginx default — means it's alive, no app ingress rules loaded yet)
```

Once app services are deployed (with ingress enabled in their Helm values), routes become active:

```bash
# Frontend
curl -s -o /dev/null -w "%{http_code}" http://<EXTERNAL-IP>/
# → 200

# API Gateway
curl -X POST http://<EXTERNAL-IP>/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","name":"Test","role":"user"}'
# → {"message":"User registered successfully..."}
```

---

## Verify ingress rules loaded

Each app service manages its own routing rule via its Helm chart values.
After deploying the apps, check that their Ingress objects exist:

```bash
kubectl get ingress -n prod
```

Expected:
```
NAME               CLASS   HOSTS   ADDRESS                                          PORTS
api-gateway        nginx   *       abc123.eu-central-1.elb.amazonaws.com            80
frontend-service   nginx   *       abc123.eu-central-1.elb.amazonaws.com            80
```

---

## Uninstallation

> ⚠️ **MUST be done BEFORE deleting the EKS cluster.**  
> If you delete the cluster first, the AWS LoadBalancer becomes an orphan — it keeps billing
> you ~$20/month with no way for Kubernetes to remove it automatically.

```bash
# Step 1: Uninstall the controller
#   This deletes the K8s Service → Cloud Controller Manager calls AWS API → LB deleted
helm uninstall ingress-nginx -n ingress-nginx

# Step 2: Wait ~30 seconds, then confirm the LB is gone from AWS
aws elbv2 describe-load-balancers \
  --region eu-central-1 \
  --query 'LoadBalancers[*].{Name:LoadBalancerName,DNS:DNSName,State:State.Code}' \
  --output table
# Should return an empty table (no rows)

# Step 3: Now safe to scale down nodes or delete the cluster
eksctl scale nodegroup --cluster=go-microservices --name=ng-1 --nodes=0 --region=eu-central-1
# or full cluster delete:
eksctl delete cluster -f infra/eks/cluster.yaml
```

---

## Current state (as of March 2026)

| Item | Value |
|---|---|
| Controller namespace | `ingress-nginx` |
| AWS LB DNS | `a12ffaa748e4147658b56ac347ce2726-1854249424.eu-central-1.elb.amazonaws.com` |
| LB IPs | `52.57.55.201`, `18.159.58.27` |
| Region | `eu-central-1` |
| Ingress rules | `api-gateway` → `/api/*`, `frontend-service` → `/` |
| App namespace | `prod` |

---

## Troubleshooting

**EXTERNAL-IP stuck as `<pending>` for more than 3 minutes:**
```bash
kubectl describe svc ingress-nginx-controller -n ingress-nginx
# Look for Events — usually a subnet tag or IAM permission issue
```

**404 on all routes after install:**
```bash
# Normal if app services haven't been deployed yet.
# Once deployed, check ingress objects exist:
kubectl get ingress -n prod
```

**LB orphaned (cluster deleted before uninstall):**
```bash
# Go to AWS Console → EC2 → Load Balancers
# Find the LB (name starts with the cluster's VPC ID)
# Delete it manually
# Also check: AWS Console → EC2 → Target Groups (delete related ones)
```

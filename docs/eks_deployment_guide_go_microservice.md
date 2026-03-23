# EKS Deployment Guide (Go Microservice)

This document is a **step-by-step reference** for deploying a Go microservice to **AWS EKS**, exactly following the flow we used (from local minikube to EKS).

---

## 0. Preconditions

* AWS account with sufficient permissions (IAM user or role)
* Docker image already built and pushed to **Amazon ECR**
* Go service with:

  * Health checks (liveness/readiness)
  * Helm chart
  * Working locally (Docker / minikube)

---

## 1. Required Tools (Local Machine)

Verify installations:

```bash
aws --version          # AWS CLI v2 (must be recent)
kubectl version --client
helm version
eksctl version
```

⚠️ **Important**: Old AWS CLI versions generate invalid kubeconfig auth (`v1alpha1`). Always use **AWS CLI v2 (latest)**.

---

## 2. AWS Authentication Check

```bash
aws sts get-caller-identity
aws configure get region
```

Set region if needed:

```bash
aws configure set region eu-central-1
```

---

## 3. Create EKS Cluster 

```bash
eksctl create cluster \
  --name go-microservices \
  --region eu-central-1 \
  --nodes 2 \
  --node-type t3.small \
  --managed
```



# 3.1 EKS Cluster with yaml

This cluster is created using `eksctl` on top of an existing VPC
provisioned via Terraform (`infra/vpc`).

```bash
#create
eksctl create cluster -f cluster.yaml

#Delete
eksctl delete cluster -f cluster.yaml

```
What this does:

* Creates EKS control plane
* Creates managed EC2 node group
* Sets up VPC, subnets, security groups
* Installs core addons (coredns, kube-proxy, vpc-cni, metrics-server)

---

## 4. Configure kubeconfig for EKS

Add context form aws
```bash
aws eks update-kubeconfig \
  --region eu-central-1 \
  --name go-microservices \
  --alias go-microservices
```

Switch context:

```bash
kubectl config use-context go-microservices
```

```bash
kubectl config use-context minikube
```

Switch back to minikube context anytime with the command above. List all available contexts:

```bash
kubectl config get-contexts
```

Delete old contexts 
```bash
kubectl config delete-context arn:aws:eks:...:old-cluster
```

All the config about the contexts are here
```bash
~/.kube/config
```
Verify connection:

```bash
kubectl get nodes -o wide
kubectl get pods -A
```

Indicators you are on **EKS**:

* Node names like `ip-192-168-xx.eu-central-1.compute.internal`
* OS: Amazon Linux

---

## 5. MongoDB Strategy (Phase 1)

For simplicity and focus on Kubernetes:

* Use **MongoDB Atlas** instead of running Mongo inside the cluster
* Temporarily allow access in Atlas:

  * Network Access → IP Allowlist → `0.0.0.0/0` (temporary)

> This avoids StatefulSets, PVCs, backups, and HA complexity during learning phase.

---

## 5.5 AWS ECR Authentication (Image Pull Access)

Since your Docker image is stored in **AWS ECR**, Kubernetes needs credentials to pull it. Follow these steps:

### Step 1: Authenticate Docker with ECR (local machine)

```bash
aws ecr get-login-password --region eu-central-1 | docker login \
  --username AWS \
  --password-stdin 114851843413.dkr.ecr.eu-central-1.amazonaws.com
```

**What this does:** Gets a temporary token from AWS and authenticates your local Docker daemon with ECR.

✔ Should output: `Login Succeeded`

### Step 2: Create Kubernetes Secret for ECR (in EKS cluster)

```bash
kubectl create secret docker-registry ecr-secret \
  --docker-server=114851843413.dkr.ecr.eu-central-1.amazonaws.com \
  --docker-username=AWS \
  --docker-password=$(aws ecr get-login-password --region eu-central-1) \
  -n default
```

**What this does:** Creates a Kubernetes Secret named `ecr-secret` that stores ECR credentials. Kubernetes uses this to authenticate when pulling images.

⚠️ **Note:** The ECR password token expires after 12 hours. You may need to recreate this secret periodically, or use **IRSA (IAM Roles for Service Accounts)** for production.

### Step 3: Add imagePullSecrets to your Helm values

In `charts/user-service/values.yaml`:

```yaml
image:
  repository: 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service
  pullPolicy: IfNotPresent
  tag: "latest"

imagePullSecrets:
  - name: ecr-secret
```

**What this does:** Tells Kubernetes to use the `ecr-secret` Secret when pulling the image.

### Step 4: Verify the credentials work

Deploy with Helm and check if the image is pulled successfully:

```bash
helm upgrade --install user-service ./charts/user-service \
  -f charts/user-service/values.yaml \
  -n default

# Check pod status
kubectl get pods -n default
kubectl describe pod <pod-name> -n default
```

✔ Pod should reach `Running` state (not `ImagePullBackOff`)

### Troubleshooting: ImagePullBackOff Error

If pods are stuck in `ImagePullBackOff`:

```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>  # May be empty if image failed to pull

# Check if the image exists in ECR
aws ecr describe-images --repository-name go-microservice/user-service --region eu-central-1
```

Common causes:

* ECR credentials expired → Recreate the secret
* Wrong ECR URI in values.yaml
* Image not pushed to ECR yet
* Wrong AWS account/region

---

## 6. Helm Values (Initial Setup)

Example key settings:

```yaml
replicaCount: 2

image:
  repository: <account>.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service
  tag: latest
  pullPolicy: Always

service:
  type: ClusterIP
  port: 50051

resources:
  requests:
    cpu: "200m"
    memory: "256Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

⚠️ Secrets were temporarily kept in plain text **locally only** (never committed).

---

## 7. Deploy Services to EKS

All services are deployed to the **`prod`** namespace. Use the `--atomic` flag so Helm auto-rolls back if the deployment fails:

```bash
# Create the namespace once
kubectl create namespace prod

# Deploy each service
helm upgrade --install user-service    ./charts/user-service    -n prod --atomic --timeout 5m -f ./charts/user-service/values.yaml    -f ./secrets/user-service.values.local.yaml
helm upgrade --install auth-service    ./charts/auth-service    -n prod --atomic --timeout 5m -f ./charts/auth-service/values.yaml    -f ./secrets/auth-service.values.local.yaml
helm upgrade --install product-service ./charts/product-service -n prod --atomic --timeout 5m -f ./charts/product-service/values.yaml -f ./secrets/product-service.values.local.yaml
helm upgrade --install api-gateway     ./charts/api-gateway     -n prod --atomic --timeout 5m -f ./charts/api-gateway/values.yaml
helm upgrade --install frontend-service ./charts/frontend-service -n prod --atomic --timeout 5m -f ./charts/frontend-service/values.yaml
```

Verify all pods are running:

```bash
kubectl get pods -n prod
kubectl get svc  -n prod
kubectl logs -l app.kubernetes.io/instance=user-service -n prod --tail=50
```

All pods should reach `1/1 Running` state.

> **In practice, all 5 services are deployed automatically by GitHub Actions CI/CD on every push to `main`. See `.github/workflows/` for the pipeline definitions.**

---

## 8. Test Service Without LoadBalancer

Use port-forwarding (same as minikube):

```bash
kubectl port-forward svc/user-service 50051:50051
```

From local machine:

```bash
grpcurl -plaintext localhost:50051 list
```

✔ Confirms:

* Image pulled from ECR
* Pod running on EKS
* MongoDB Atlas connectivity works

---

## 9. Ingress Controller & Public Access

### 9.1 Why port-forward is not enough

`kubectl port-forward` drills a temporary tunnel from your laptop to a pod. It stops when you close the terminal, only you can access it, and it is not how production works.

With only `ClusterIP` services (what all our services use), the cluster looks like this:

```
Internet
   ❌ (no way in)

EKS Cluster
├── api-gateway   ClusterIP 172.20.x.x:8080   ← internal only
├── auth-service  ClusterIP 172.20.x.x:50052  ← internal only
├── user-service  ClusterIP 172.20.x.x:50051  ← internal only
└── frontend      ClusterIP 172.20.x.x:80     ← internal only
```

### 9.2 How Ingress works

```
Internet
   ↓  HTTP/HTTPS
AWS LoadBalancer  ← one real public DNS (e.g. abc123.eu-central-1.elb.amazonaws.com)
   ↓
nginx Ingress Controller (a pod inside your cluster)
   ↓  routes based on path rules
   ├── /api/*  → api-gateway:8080
   └── /*      → frontend-service:80
```

Three concepts to keep separate:

| Concept | What it is |
|---|---|
| **Ingress resource** | A YAML file with routing rules (path → service) |
| **Ingress Controller** | An nginx pod that reads those rules and does the actual routing |
| **AWS LoadBalancer** | The AWS resource that forwards internet traffic into the nginx pod |

### 9.3 How Kubernetes creates the AWS LoadBalancer without Terraform

This is the part that seems like magic. When you created the EKS cluster with `eksctl`, it automatically:

1. **Created an IAM role** for the node group with permissions like `elasticloadbalancing:CreateLoadBalancer`, `ec2:DescribeSubnets`, etc.
2. **Installed the AWS Cloud Controller Manager** — a pod running inside your cluster that watches Kubernetes for `Service type: LoadBalancer` objects and calls the AWS API on your behalf.

So the full flow when you run `helm install ingress-nginx`:

```
helm install ingress-nginx ...
      ↓
Helm creates a Kubernetes Service with type: LoadBalancer
      ↓
AWS Cloud Controller Manager sees it (watching the K8s API)
      ↓
It calls AWS API: "CreateLoadBalancer" in eu-central-1
      ↓
AWS creates a Classic/NLB Load Balancer in EC2
      ↓
Assigns public DNS: abc123.eu-central-1.elb.amazonaws.com
      ↓
Writes that DNS back into the Service's .status.loadBalancer.ingress field
```

**Terraform vs Helm/K8s for infrastructure:**

| | Terraform | Helm/Kubernetes |
|---|---|---|
| **Who calls AWS API** | Terraform CLI on your machine | AWS Cloud Controller inside cluster |
| **State tracked in** | `terraform.tfstate` | Kubernetes objects |
| **Deleted when** | `terraform destroy` | `helm uninstall` or `kubectl delete svc` |
| **Best for** | VPC, subnets, IAM, EKS cluster itself | Everything *inside* the cluster |

The LoadBalancer created by Helm appears in **AWS Console → EC2 → Load Balancers**, not in Terraform state.

### 9.4 The three-layer mental model

```
Layer 1 — AWS Infrastructure (Terraform / eksctl)  — changes rarely
  VPC, Subnets, Security Groups, EKS Cluster, Node IAM Role

Layer 2 — Cluster-wide tools (Helm, installed once)
  nginx Ingress Controller, cert-manager, Prometheus...
  These talk to AWS via the node IAM role

Layer 3 — Your apps (CI/CD, deployed on every push)
  user-service, auth-service, api-gateway, frontend...
```

### 9.5 Install nginx Ingress Controller

> **⚠️ Node capacity check first:** A `t3.small` node has a hard limit of **11 pods**. With 5 services + 6 system pods you hit the cap. The ingress controller adds 2 more pods — you must have at least **2 nodes** before installing:
> ```bash
> # Check current pod count vs capacity
> kubectl describe node | grep -E 'Allocatable:|Non-terminated Pods:'
>
> # Scale nodegroup to 2 nodes if needed
> eksctl scale nodegroup --cluster=go-microservices --name=ng-1 --nodes=2
> ```

Add the Helm repo (once) and install:

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --wait --timeout 3m
```

Wait for AWS to provision the LoadBalancer (~2 minutes), then get the public hostname:

```bash
kubectl get svc -n ingress-nginx
# Look for EXTERNAL-IP — it will be something like:
# abc123.eu-central-1.elb.amazonaws.com
```

### 9.6 Enable Ingress via Helm chart values

> **⚠️ Do NOT use `kubectl apply` for Ingress resources.** If you create them manually, Helm doesn't know about them — they won't be updated on redeploy and won't be deleted on `helm uninstall`. Always manage Ingress through the chart.

Each service chart already has an Ingress template at `charts/<service>/templates/ingress.yaml`. You only need to enable it in `values.yaml`:

**`charts/api-gateway/values.yaml`** — routes `/api/*` to api-gateway:
```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations:
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
  hosts:
    - host: ""          # empty = wildcard, matches any hostname
      paths:
        - path: /api/
          pathType: Prefix
```

**`charts/frontend-service/values.yaml`** — routes `/` to frontend:
```yaml
ingress:
  enabled: true
  className: "nginx"
  annotations: {}
  hosts:
    - host: ""
      paths:
        - path: /
          pathType: Prefix
```

Redeploy both services to apply the ingress:

```bash
helm upgrade --install api-gateway      ./charts/api-gateway      -n prod --atomic -f ./charts/api-gateway/values.yaml
helm upgrade --install frontend-service ./charts/frontend-service  -n prod --atomic -f ./charts/frontend-service/values.yaml
```

Verify ingress resources and their ADDRESS field:

```bash
kubectl get ingress -n prod
# Both should show the elb.amazonaws.com hostname in ADDRESS
```

Expected output:
```
NAME               CLASS   HOSTS   ADDRESS                                     PORTS   AGE
api-gateway        nginx   *       abc123.eu-central-1.elb.amazonaws.com       80      1m
frontend-service   nginx   *       abc123.eu-central-1.elb.amazonaws.com       80      1m
```

Then test from your browser or curl:

```bash
ELB="<paste-your-elb-hostname-here>"

# Frontend (should return 200)
curl -s -o /dev/null -w "%{http_code}" http://$ELB/

# API Gateway — real route is /api/v1/register, not /healthz
curl -s -X POST http://$ELB/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","name":"Test","role":"user"}'
# Should return: {"message":"User registered successfully..."}
```

> **Verified:** `POST /api/v1/register` returned HTTP 200 with a user_id — meaning the full chain worked: Internet → ELB → nginx → api-gateway → auth-service → MongoDB.

### 9.7 ⚠️ Cost warning & safe shutdown order

The AWS LoadBalancer costs **~$20/month** even when idle. Since you turn the cluster on/off, always delete the ingress controller **before** deleting the cluster, otherwise the LoadBalancer becomes an orphan in AWS — still billing you, but nothing manages it.

**Safe shutdown order:**

```bash
# Step 1: Delete ingress controller (removes LoadBalancer from AWS)
helm uninstall ingress-nginx -n ingress-nginx

# Step 2: Verify LoadBalancer is gone
aws elbv2 describe-load-balancers --region eu-central-1

# Step 3: Now safe to delete the cluster
eksctl delete cluster -f infra/eks/cluster.yaml
```

---

## 10. How to Confirm You Are on EKS (Checklist)

```bash
kubectl config current-context
kubectl get nodes -o wide
kubectl cluster-info
```

Expected:

* Context is NOT `minikube`
* Control plane URL ends with `eks.amazonaws.com`

---

## 11. Known Issues & Fixes

### Error: `invalid apiVersion client.authentication.k8s.io/v1alpha1`

Cause:

* Old AWS CLI version

Fix:

* Upgrade AWS CLI v2
* Re-run `aws eks update-kubeconfig`

---

## 12. Cleanup (Delete Everything From AWS)

When you are done testing, remove AWS resources to avoid ongoing costs.

### 12.1 Delete the EKS cluster (created by eksctl)

```bash
eksctl delete cluster \
  --name go-microservices \
  --region eu-central-1
```

This typically deletes:

* EKS control plane
* Managed node group(s)
* VPC/subnets/security groups created by eksctl
* CloudFormation stacks created by eksctl

### 12.2 Verify no EKS clusters remain

```bash
aws eks list-clusters --region eu-central-1
```

### 12.3 Clean up ECR repositories (optional)

ECR repos and images are **not** removed automatically when you delete the cluster.

List repos:

```bash
aws ecr describe-repositories
```

Delete a repo (danger: deletes all images):

```bash
aws ecr delete-repository \
  --repository-name go-microservice/user-service \
  --force
```

### 12.4 Cost sanity checks (recommended)

If you created any LoadBalancers, verify they are removed:

```bash
aws elbv2 describe-load-balancers
```

Optionally check remaining EC2 instances:

```bash
aws ec2 describe-instances
```

### 12.5 Atlas reminder

If you temporarily allowed wide IP access in MongoDB Atlas (e.g., `0.0.0.0/0`), revert it after testing.

---

## 13. Recommended Next Steps (Production-Grade)

**Already completed ✅**
1. ✅ Move sensitive env vars to **Kubernetes Secrets** (Helm `secret.yaml` templates per service)
2. ✅ All 5 services deployed to `prod` namespace with Helm
3. ✅ CI/CD via **GitHub Actions** — push to `main` → ECR → `helm upgrade --atomic`
4. ✅ nginx Ingress Controller with public AWS LoadBalancer
5. ✅ Ingress rules managed by Helm charts (not manual kubectl)

**Up next:**
6. **Domain + TLS** — Register a domain, point it at the ELB, add `cert-manager` + Let's Encrypt for HTTPS
7. **Observability** — Deploy Prometheus + Grafana; add `/metrics` endpoints to Go services
8. **IRSA (IAM Roles for Service Accounts)** — Replace the 12-hour ECR token (`ecr-secret`) with a permanent IAM role bound to a K8s service account
9. **Horizontal Pod Autoscaling** — Add HPA based on CPU/memory metrics from Prometheus
10. **Network Policies** — Restrict pod-to-pod traffic (e.g. only api-gateway can call auth-service)

---

## Summary

You now have:

* All **5 Go microservices** running on **AWS EKS** (`prod` namespace)
  * user-service, auth-service, product-service, api-gateway, frontend-service
* **Helm-based deployment** with per-service charts (secrets, configmaps, ingress, health probes)
* **CI/CD pipeline** — GitHub Actions builds Docker images, pushes to ECR, deploys via `helm upgrade --atomic`
* **Stuck-release protection** — every workflow detects and cleans up `pending-*` Helm states before deploying
* **nginx Ingress Controller** routing public traffic: `/api/*` → api-gateway, `/` → frontend
* **AWS LoadBalancer** provisioned automatically by the Kubernetes Cloud Controller
* **MongoDB Atlas** for persistence (auth + user + product services)
* End-to-end verified: `POST /api/v1/register` returns HTTP 200 from the public internet

This setup is **resume-ready** and production-aligned.

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

## 7. Deploy Service to EKS

```bash
helm upgrade --install user-service ./charts/user-service -n default -f ./charts/user-service/values.yaml -f ./secrets/user-service.values.local.yaml

```

Verify:

```bash
kubectl get pods
kubectl get svc
kubectl logs -l app.kubernetes.io/instance=user-service --tail=200
kubectl get secret user-service-secrets -n default -o jsonpath='{.data}' ; echo  
```

Pod should reach `Running` state.

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

## 9. How to Confirm You Are on EKS (Checklist)

```bash
kubectl config current-context
kubectl get nodes -o wide
kubectl cluster-info
```

Expected:

* Context is NOT `minikube`
* Control plane URL ends with `eks.amazonaws.com`

---

## 10. Known Issues & Fixes

### Error: `invalid apiVersion client.authentication.k8s.io/v1alpha1`

Cause:

* Old AWS CLI version

Fix:

* Upgrade AWS CLI v2
* Re-run `aws eks update-kubeconfig`

---

## 11. Cleanup (Delete Everything From AWS)

When you are done testing, remove AWS resources to avoid ongoing costs.

### 11.1 Delete the EKS cluster (created by eksctl)

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

### 11.2 Verify no EKS clusters remain

```bash
aws eks list-clusters --region eu-central-1
```

### 11.3 Clean up ECR repositories (optional)

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

### 11.4 Cost sanity checks (recommended)

If you created any LoadBalancers, verify they are removed:

```bash
aws elbv2 describe-load-balancers
```

Optionally check remaining EC2 instances:

```bash
aws ec2 describe-instances
```

### 11.5 Atlas reminder

If you temporarily allowed wide IP access in MongoDB Atlas (e.g., `0.0.0.0/0`), revert it after testing.

---

## 12. Recommended Next Steps (Production-Grade)

1. Move sensitive env vars to **Kubernetes Secrets** (or External Secrets)
2. Enable **OIDC / IRSA** on the cluster
3. Add **LoadBalancer (NLB) for gRPC** or ALB Ingress
4. Deploy the second service using the same pattern
5. Add CI/CD (GitHub Actions → ECR → Helm deploy)
6. Add observability (logs, metrics)

---

## Summary

You now have:

* Go microservice running on **AWS EKS**
* Helm-based deployment
* MongoDB Atlas integration
* Clean separation between local (minikube) and cloud (EKS)

This setup is **resume-ready** and production-aligned.

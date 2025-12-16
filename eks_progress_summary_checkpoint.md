# EKS Progress Summary – Checkpoint

This file is a **concise checkpoint** of everything completed so far, intended to help resume work without re‑deriving context.

---

## Goal
Deploy Go microservices to **AWS EKS** in a production‑grade, repeatable way suitable for resume/interview discussion.

---

## What Is Completed ✅

### 1. Tooling & Access
- AWS CLI **v2 upgraded** (fixed kubeconfig `v1alpha1` auth error)
- kubectl, Helm, eksctl installed
- AWS IAM access verified with `aws sts get-caller-identity`

---

### 2. EKS Lifecycle (Create / Delete)
- EKS cluster created using **eksctl** (managed node group)
- kubeconfig generated with `aws eks update-kubeconfig`
- Context switching between minikube and EKS understood and validated
- Full cleanup tested:
  - `eksctl delete cluster`
  - ECR intentionally preserved

---

### 3. Image & Registry
- Docker image built and pushed to **Amazon ECR**
- Fixed `ImagePullBackOff` caused by incorrect `image.repository` (DockerHub fallback)
- Verified EKS nodes can pull from ECR

---

### 4. user-service Deployment (Helm)
- `user-service` deployed successfully to EKS
- gRPC service running and tested via `kubectl port-forward`
- Health probes (liveness/readiness) working

---

### 5. MongoDB Strategy
- MongoDB **Atlas** used instead of in‑cluster Mongo
- Connectivity validated from EKS
- Atlas IP allowlist temporarily relaxed for testing

---

### 6. Secrets Handling (Important Milestone)

#### Structure
- Created local directory: `secrets/` (gitignored)
- Per‑service local values file:
  - `secrets/user-service.values.local.yaml`

#### Helm Integration
- Added `templates/secret.yaml` to chart
- Secrets injected via:
  ```yaml
  envFrom:
    - secretRef:
        name: user-service-secrets
  ```
- Non‑sensitive envs (`PORT`, `MONGO_DB`) kept in `env:`

#### Validation
- Kubernetes Secret exists and contains expected keys
- Pod runs correctly even though `kubectl describe pod` does not list `envFrom` vars explicitly
- RabbitMQ intentionally excluded for now

---

### 7. Debugging Skills Applied
- kubeconfig auth version mismatch
- stale kubeconfig contexts
- ECR vs DockerHub image resolution
- Secret key name mismatches
- Containers without `sh`

---

## Current State 🟢
- `user-service` runs correctly on **EKS**
- Helm chart is clean and reusable
- Secrets are handled safely (local only, not committed)
- Deployment pattern is repeatable

---

## Agreed Next Steps 🔜

1. Deploy **second service** using the same Helm + Secret pattern
2. Separate environment values (`values.eks.yaml`, etc.)
3. (Later) Introduce:
   - IRSA
   - AWS Secrets Manager + External Secrets Operator
   - LoadBalancer / NLB for gRPC
   - CI/CD (GitHub Actions → ECR → Helm)

---

## Key Resume Bullet (Draft)
> Deployed Go microservices on AWS EKS using Helm, eksctl, and Amazon ECR; implemented per‑service secret management, MongoDB Atlas integration, health probes, and repeatable cluster lifecycle management.

---

**Status:** Ready to proceed with service #2 🚀


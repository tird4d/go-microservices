# 🚀 GitHub Actions → Kubernetes CI/CD Guide

**Your Goal:** When you push code → GitHub Actions builds image → Pushes to ECR → Deploys to EKS with Helm

---

## 📊 Your Current Setup vs New Setup

### **CURRENT: Docker-Compose + SSH**
```
git push production
    ↓
GitHub Actions triggered
    ↓
SSH into EC2 server
    ↓
git pull on server
    ↓
docker-compose build
    ↓
docker-compose up -d
```

**Advantages:**
- Simple, all happens on one server
- Easy to debug (just SSH in)

**Disadvantages:**
- Can't scale
- Server must always be running
- Manual server management

---

### **NEW: Kubernetes + ECR**
```
git push main
    ↓
GitHub Actions triggered
    ↓
Build Docker image
    ↓
Push to ECR (AWS)
    ↓
kubectl/helm connects to EKS
    ↓
Deploy new image
    ↓
Kubernetes handles scaling, health, rollback
```

**Advantages:**
- Scalable, automated
- Cloud-native
- Self-healing (restarts failed pods)
- Easy rollback

**Disadvantages:**
- More moving parts
- Need AWS credentials in GitHub

---

## 🔑 Key Differences You Need to Understand

### **1. Authentication**

**Your current approach (Docker-Compose):**
```yaml
- name: SSH into server
  uses: appleboy/ssh-action@master
  with:
    host: ${{ secrets.SSH_HOST }}
    key: ${{ secrets.SSH_KEY }}  # Private key to server
    script: docker-compose up -d
```
- You authenticate to a **server** (SSH key)
- Commands run ON the server

**New approach (Kubernetes):**
```yaml
- name: Configure AWS credentials
  uses: aws-actions/configure-aws-credentials@v4
  with:
    aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY }}
    aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
    aws-region: eu-central-1

- name: Push to ECR
  run: |
    aws ecr get-login-password --region eu-central-1 | \
    docker login --username AWS --password-stdin 114851843413.dkr.ecr.eu-central-1.amazonaws.com
    docker push 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:latest
```
- You authenticate to **AWS** (access keys)
- You authenticate to **Kubernetes** (kubeconfig)
- Commands run in GitHub Actions runner (cloud)

---

### **2. Image Registry**

**Your current approach:**
- No registry needed (all on server)
- `docker-compose build` creates images locally

**New approach:**
```
GitHub Actions → Build image → Push to ECR → EKS pulls from ECR
                                    ↓
                              (Docker image stored in cloud)
```

**Why ECR?**
- EKS has permissions to pull from ECR
- Image is available globally
- Survives if pod needs to restart

---

### **3. Deployment Method**

**Your current approach:**
```bash
docker-compose up -d
```
- Simple command
- All services in one compose file
- Manual updates

**New approach:**
```bash
helm upgrade --install user-service ./charts/user-service \
  -f values.yaml \
  -f secrets/user-service.values.local.yaml
```
- Uses Helm (templating system)
- Can deploy multiple services independently
- Handles rollback automatically
- Supports health checks, scaling, etc.

---

## 🏗️ GitHub Actions Workflow Structure for Kubernetes

### **Complete Flow:**

```yaml
name: Build and Deploy User Service to EKS

on:
  push:
    branches: [main]
    paths:
      - 'user_service/**'  # Only trigger if user_service code changed
      - 'charts/user-service/**'
      - '.github/workflows/user-service-deploy.yml'

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    
    steps:
      # Step 1: Get the code
      - name: Checkout code
        uses: actions/checkout@v3
      
      # Step 2: Configure AWS credentials
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: eu-central-1
      
      # Step 3: Login to ECR
      - name: Login to Amazon ECR
        run: |
          aws ecr get-login-password --region eu-central-1 | \
          docker login --username AWS --password-stdin 114851843413.dkr.ecr.eu-central-1.amazonaws.com
      
      # Step 4: Build Docker image
      - name: Build Docker image
        run: |
          cd user_service
          docker build -t 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:${{ github.sha }} .
          docker tag 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:${{ github.sha }} \
                     114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:latest
      
      # Step 5: Push to ECR
      - name: Push to Amazon ECR
        run: |
          docker push 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:${{ github.sha }}
          docker push 114851843413.dkr.ecr.eu-central-1.amazonaws.com/go-microservice/user-service:latest
      
      # Step 6: Update kubeconfig
      - name: Update kubeconfig
        run: |
          aws eks update-kubeconfig \
            --region eu-central-1 \
            --name go-microservices
      
      # Step 7: Deploy with Helm
      - name: Deploy with Helm
        run: |
          helm upgrade --install user-service ./charts/user-service \
            -f charts/user-service/values.yaml \
            -f secrets/user-service.values.local.yaml \
            --wait \
            --timeout 5m
      
      # Step 8: Verify deployment
      - name: Verify deployment
        run: |
          kubectl rollout status deployment/user-service -n default --timeout=5m
          kubectl get pods -l app.kubernetes.io/instance=user-service
```

---

## 🔐 AWS Credentials Setup

### **Step 1: Create GitHub Secrets**

You need to add these to your GitHub repo:

Settings → Secrets and variables → Actions → New repository secret

Add:
```
AWS_ACCESS_KEY_ID          = (from AWS IAM user)
AWS_SECRET_ACCESS_KEY      = (from AWS IAM user)
```

### **Step 2: Get AWS Credentials**

In AWS Console:
1. IAM → Users → Create user (e.g., `github-actions`)
2. Attach policy: `AmazonEC2ContainerRegistryPowerUser` + `EKSServiceRolePolicy`
3. Create access keys
4. Copy to GitHub Secrets

### **Step 3: Test Locally**

```bash
export AWS_ACCESS_KEY_ID="your_key"
export AWS_SECRET_ACCESS_KEY="your_secret"
aws ecr describe-repositories
```

If this works, GitHub Actions will work too.

---

## 📝 Key Concepts Explained

### **1. github.sha Variable**
```yaml
docker build -t image:${{ github.sha }}
```
- Unique identifier for each commit
- Example: `image:a1b2c3d4e5f6...`
- Used for image versioning
- Can rollback to any previous commit

### **2. Push vs Deployment Trigger**

```yaml
on:
  push:
    branches: [main]
    paths:
      - 'user_service/**'  # Only trigger if these files change
```

**Example:**
- Change `user_service/main.go` → Trigger
- Change `product_service/main.go` → Don't trigger
- Change `README.md` → Don't trigger

### **3. Helm --wait Flag**

```yaml
helm upgrade --install user-service ./charts/user-service \
  --wait \
  --timeout 5m
```

- `--wait` = Don't return until pods are ready
- `--timeout 5m` = Wait max 5 minutes

### **4. Rollout Status Check**

```bash
kubectl rollout status deployment/user-service
```

- Verifies the deployment actually succeeded
- Checks if pods became ready
- Fails if timeout

---

## ⚠️ Common Mistakes to Avoid

### ❌ **Mistake 1: Hardcoding AWS Account ID**
```yaml
docker push 114851843413.dkr.ecr.eu-central-1.amazonaws.com/...
```
This breaks if you change AWS accounts.

**Better:**
```bash
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
docker push ${AWS_ACCOUNT_ID}.dkr.ecr.eu-central-1.amazonaws.com/...
```

### ❌ **Mistake 2: Not checking if deployment succeeded**
```yaml
helm upgrade ...  # Just runs, doesn't verify
```

**Better:**
```yaml
helm upgrade ...
kubectl rollout status deployment/user-service
```

### ❌ **Mistake 3: Using `latest` tag without commit SHA**
```yaml
docker push image:latest  # Can't rollback
```

**Better:**
```yaml
docker push image:${{ github.sha }}
docker push image:latest  # Also tag as latest
```

### ❌ **Mistake 4: Not triggering on chart changes**
```yaml
on:
  push:
    paths:
      - 'user_service/**'  # Missing Helm chart!
```

**Better:**
```yaml
on:
  push:
    paths:
      - 'user_service/**'
      - 'charts/user-service/**'  # Re-deploy if chart changes
      - '.github/workflows/user-service-deploy.yml'
```

---

## 🎯 What You Need to Build

Now it's YOUR turn! Create file: `.github/workflows/user-service-deploy.yml`

### **TODO:**

1. **Copy the template above** into `.github/workflows/user-service-deploy.yml`
2. **Replace these placeholders:**
   - `114851843413` → Your AWS account ID (check in secrets file)
   - `eu-central-1` → Your AWS region
   - `go-microservice/user-service` → Your ECR repo name
   - `user-service` → Service name
   - `./charts/user-service` → Path to your Helm chart

3. **Create GitHub Secrets:**
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`

4. **Test:**
   - Push code to main
   - Watch GitHub Actions run
   - Verify pod deployed to EKS

---

## ✅ Verification Checklist

After creating workflow:

```bash
# 1. Is the workflow file valid?
cat .github/workflows/user-service-deploy.yml

# 2. Are GitHub Secrets set?
# Check: Settings → Secrets → Should see AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY

# 3. Turn on EKS cluster
eksctl create cluster -f infra/eks/cluster.yaml

# 4. Push code
git add .github/workflows/user-service-deploy.yml
git commit -m "Add GitHub Actions CI/CD for user-service"
git push origin main

# 5. Watch GitHub Actions
# Go to: GitHub Repo → Actions → Latest run

# 6. Verify deployment
kubectl get pods
kubectl logs -l app.kubernetes.io/instance=user-service
```

---

## 🚀 Your Task

**Create the GitHub Actions workflow file** using the template above.

Questions before you start?
- Need help understanding any step?
- Unsure where to find your AWS account ID?
- Want me to review before you push?

Let me know when you've created the file! 🎯

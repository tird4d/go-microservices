# User Service Deployment with Helm & Kubernetes

This project demonstrates deploying a Go-based `user-service` using Docker, Kubernetes (via Minikube), and Helm.

---

## 📁 Folder Structure

```
user-service/
├── charts/
│   └── user-service/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── deployment.yaml
│           ├── service.yaml
│           └── ...
├── main.go
├── Dockerfile
├── go.mod
├── go.sum
├── proto/
└── ...
```

---

## 🚀 Steps to Build & Deploy

### 1. Enable Minikube Docker Env
```bash
eval $(minikube docker-env) 
```
#### 1.2. Or after creating image copy that to mini kube
```bash
docker minikube image load user-service:latest
```

### 2. Build Docker Image
```bash
docker build -t user-service:latest .
```

### 3. Create Helm Chart
Make sure you have a Helm chart inside `charts/user-service/`.

You can generate it via:
```bash
helm create charts/user-service
```

Then **replace the generated files** with your own `deployment.yaml`, `service.yaml`, and update `values.yaml`.

### 4. Load Docker Image (Optional)
You only need this if Minikube can't see your local image.
```bash
minikube image load user-service:latest
```

### 5. Deploy via Helm
```bash
helm upgrade --install user-service ./charts/user-service
```

---

## 🧪 Test Your Service

### Port forward for local testing:
```bash
kubectl port-forward deployments/user-service 50051:50051
```

Then test your gRPC service or use Postman/gRPC UI.

---

## 🔧 Troubleshooting

- If `ImagePullBackOff`: make sure you ran `eval $(minikube docker-env)` before building.
- If service not found: check `kubectl get pods` and logs with:
```bash
kubectl logs <pod-name>
```

---

## ✅ Done

Your Go service is now deployed on Kubernetes via Helm!



kubectl get svc
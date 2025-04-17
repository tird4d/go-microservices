# 🐳 user-service Deployment with Kubernetes & Helm

This guide describes how to build, package, and deploy the `user-service` using Docker, Minikube, and Helm.

---

## 📦 Project Structure

```
/user-service
├── Dockerfile
├── go.mod / go.sum / vendor/
├── main.go
├── .env
├── handlers/, config/, services/, ...
├── Chart.yaml
├── values.yaml
└── templates/         ← Helm Templates
```

---

## ⚙️ Requirements

- Docker
- Minikube
- Helm

---

## 🚀 Step-by-Step Deployment Guide

### 1. Enable Docker for Minikube
```bash
eval $(minikube docker-env)
```

---

### 2. Build the Docker Image
```bash
docker build -t user-service:latest .
```

Check the image:
```bash
docker images
```

---

### 3. Deploy with Helm

If you're already inside the `user-service` folder:
```bash
helm upgrade --install user-service .
```

Or if the Helm Chart is inside a subfolder:
```bash
helm upgrade --install user-service charts/user-service
```

---

### 4. Check Pod Status

```bash
kubectl get pods
kubectl logs deployment/user-service
```

---

### 5. Port Forward (gRPC)

```bash
kubectl port-forward deployment/user-service 50051:50051
```

---

## 🧠 Notes

- Use `vendor/` to ensure all dependencies are inside the image.
- If your code uses private modules, clone them locally and use `replace` in `go.mod`.
- Use `service name` (not `localhost`) when connecting between services in Kubernetes.
- You can use a `.env` file and load it in your Go code using `github.com/joho/godotenv`.

---

## 🔧 Common Issues

| Problem                | Reason                                 | Solution                                  |
|------------------------|----------------------------------------|-------------------------------------------|
| `ImagePullBackOff`     | Image not found in Minikube            | Use `eval $(minikube docker-env)` before build |
| `connection refused`   | Port is not forwarded or wrong         | Check `kubectl port-forward`, check logs  |
| Env vars not loaded    | `.env` not used properly in Go code    | Use `godotenv.Load()` in `main.go`        |

---

## ✅ Done!

Your user-service is now running inside Kubernetes and accessible via port `50051`.
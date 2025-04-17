
# مستندسازی پروژه Microservices با Go، gRPC، و Kubernetes

این فایل توضیحاتی درباره نحوه ساخت، اجرای لوکال، و دیپلوی سرویس‌های این پروژه با استفاده از Docker، Helm و Minikube را فراهم می‌کند.

## ساختار پوشه‌ها

```
user-service/
├── Dockerfile
├── go.mod
├── main.go
├── ...
└── charts/
    └── user-service/
        ├── Chart.yaml
        ├── values.yaml
        └── templates/
            ├── deployment.yaml
            └── service.yaml
```

## مراحل اجرای پروژه

### 1. تنظیم محیط Docker برای Minikube

```bash
eval $(minikube docker-env)
```

### 2. ساخت image برای سرویس

```bash
docker build -t user-service:latest .
```

> توجه: حتما باید داخل دایرکتوری `user-service/` این دستور اجرا شود.

### 3. اجرای Helm برای دیپلوی سرویس

```bash
helm upgrade --install user-service charts/user-service
```

### 4. بررسی وضعیت پادها

```bash
kubectl get pods
```

### 5. دسترسی به سرویس

```bash
kubectl port-forward deployment/user-service 50051:50051
```

## نکات مهم

- اگر `ImagePullBackOff` دریافت کردید، اطمینان حاصل کنید image به درستی ساخته شده باشد.
- اگر از `replace` در `go.mod` استفاده شده، حتما از `vendor/` استفاده کنید.
- از `values.yaml` برای تعریف envها، نام ایمیج، و سایر تنظیمات استفاده کنید.
- در صورت نیاز به تغییر در مسیرها یا تنظیمات، حتما `helm upgrade` را مجدد اجرا کنید.

---

برای افزودن سرویس جدید، کافیست مراحل مشابه را با نام جدید دنبال کنید.

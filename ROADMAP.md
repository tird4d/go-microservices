# 🗺️ Roadmap: از صفر تا Production

> **هدف:** یادگیری Go + Microservices + Kubernetes + Cloud Deployment
> 
> **زمان تخمینی:** ۸-۱۲ هفته (با کار روزانه ۱-۲ ساعت)

---

## 📍 وضعیت فعلی تو

براساس فایل‌های پروژه:
- ✅ ساختار پایه microservices رو داری
- ✅ Docker و docker-compose کار میکنه
- ✅ gRPC بین سرویس‌ها پیاده شده
- ✅ RabbitMQ برای messaging
- ✅ MongoDB و Redis
- ✅ JWT authentication
- ⏸️ Kubernetes (شروع شده ولی ناقص)
- ❌ Cloud deployment

---

## 🎯 فاز ۱: یادآوری و اجرای مجدد (هفته ۱)

### روز ۱-۲: محیط و اجرا
- [ ] پروژه رو با `docker-compose up` اجرا کن
- [ ] مطمئن شو همه سرویس‌ها بالا میان
- [ ] از Postman/curl یه request بزن و تست کن

### روز ۳-۴: مرور کد Go
- [ ] فایل `main.go` هر سرویس رو بخون
- [ ] ساختار handler ها رو مرور کن
- [ ] نحوه اتصال به MongoDB/Redis رو ببین

### روز ۵-۷: مرور gRPC
- [ ] فایل‌های `.proto` رو بخون
- [ ] نحوه generate کردن کد رو یادت بیار
- [ ] یه endpoint جدید اضافه کن (برای تمرین)

**چک‌لیست تکمیل فاز ۱:**
```bash
# این دستورات باید کار کنن
cd project && make up_build
curl http://localhost:8080/health
```

---

## 🎯 فاز ۲: تکمیل و بهبود کد (هفته ۲-۳)

### Structured Logging
- [ ] نصب `zap` logger
- [ ] جایگزینی `log.Println` با `zap`
- [ ] اضافه کردن context به لاگ‌ها

### Error Handling
- [ ] ایجاد custom error types
- [ ] پیاده‌سازی error middleware
- [ ] برگرداندن error codes مناسب

### Health Checks
- [ ] اضافه کردن `/health` endpoint به همه سرویس‌ها
- [ ] چک کردن اتصال به dependencies

### Graceful Shutdown
- [ ] Handle کردن SIGTERM/SIGINT
- [ ] بستن صحیح connections

**خروجی فاز ۲:**
- کد تمیزتر و قابل debug
- لاگ‌های ساختاریافته
- سرویس‌های stable

---

## 🎯 فاز ۳: Testing (هفته ۴)

### Unit Tests
- [ ] نوشتن تست برای business logic
- [ ] استفاده از `testify` برای assertions
- [ ] Mock کردن dependencies

### Integration Tests
- [ ] تست endpoint های gRPC
- [ ] تست با database واقعی

### اجرای تست‌ها
```bash
go test ./... -v
go test ./... -cover
```

**خروجی فاز ۳:**
- حداقل ۶۰% code coverage
- CI-ready test suite

---

## 🎯 فاز ۴: Kubernetes Local (هفته ۵-۶)

### Setup
- [ ] نصب Minikube
- [ ] نصب kubectl
- [ ] نصب Helm

### یادگیری مفاهیم
- [ ] Pod, Deployment, Service
- [ ] ConfigMap, Secret
- [ ] Ingress
- [ ] Probes (liveness, readiness)

### دیپلوی سرویس‌ها
- [ ] ایجاد Helm chart برای هر سرویس
- [ ] دیپلوی روی Minikube
- [ ] تست ارتباط بین سرویس‌ها

```bash
minikube start
eval $(minikube docker-env)
helm upgrade --install broker-service ./charts/broker-service
```

**خروجی فاز ۴:**
- همه سرویس‌ها روی K8s local اجرا میشن
- Ingress کار میکنه

---

## 🎯 فاز ۵: Observability (هفته ۷)

### Metrics (Prometheus)
- [ ] اضافه کردن metrics endpoint به سرویس‌ها
- [ ] نصب Prometheus با Helm
- [ ] تعریف alerting rules

### Dashboards (Grafana)
- [ ] نصب Grafana
- [ ] ساخت dashboard برای هر سرویس
- [ ] نمایش request count, latency, errors

### Tracing (اختیاری)
- [ ] نصب Jaeger
- [ ] پیاده‌سازی distributed tracing

**خروجی فاز ۵:**
- داشبورد monitoring
- آلرت برای مشکلات

---

## 🎯 فاز ۶: CI/CD (هفته ۸)

### GitHub Actions
- [ ] ایجاد workflow برای test
- [ ] ایجاد workflow برای build
- [ ] Push به container registry

### مثال workflow:
```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test ./...
```

**خروجی فاز ۶:**
- Automated testing
- Automated Docker builds

---

## 🎯 فاز ۷: Cloud Deployment (هفته ۹-۱۲)

### انتخاب Cloud Provider
| Provider | سرویس K8s | مزیت |
|----------|-----------|------|
| AWS | EKS | بازار کار بیشتر |
| Azure | AKS | ساده‌تر برای شروع |
| GCP | GKE | بهترین K8s experience |

### مراحل
- [ ] ایجاد اکانت cloud
- [ ] راه‌اندازی cluster
- [ ] Setup container registry
- [ ] Deploy با Helm
- [ ] Configure Ingress + TLS
- [ ] Setup domain

### Infrastructure as Code
- [ ] یادگیری Terraform basics
- [ ] نوشتن terraform برای cluster

**خروجی نهایی:**
- ✅ پروژه live روی cloud
- ✅ HTTPS با domain واقعی
- ✅ Monitoring فعال
- ✅ CI/CD کامل

---

## 📚 منابع یادگیری

### Go
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

### Kubernetes
- [Kubernetes Basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
- [Helm Documentation](https://helm.sh/docs/)

### gRPC
- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/basics/)

---

## 🚦 شروع کن!

**قدم اول همین الان:**
```bash
cd /home/tirdad/Projects/go-microservices/project
docker-compose up -d
docker ps
```

اگه همه چی بالا اومد، فاز ۱ رو شروع کردی! 🎉

---

> 💡 **نکته:** هر هفته این فایل رو آپدیت کن و پیشرفتت رو track کن.
> 
> ❓ **سوال داری؟** هر مرحله که گیر کردی، بپرس!

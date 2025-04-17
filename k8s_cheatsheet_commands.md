## 📦 Deployment Commands (Helm, kubectl, minikube)

### 🔁 Helm

```bash
# نصب یا آپدیت سرویس
helm upgrade --install <release-name> <chart-path>

# مشاهده لیست release ها
helm list

# حذف یک release
helm uninstall <release-name>

# بررسی render خروجی yaml نهایی
helm template <release-name> <chart-path>

# مشاهده فایل‌های rendered و diff با deployment قبلی
helm diff upgrade <release-name> <chart-path>   # نیاز به نصب افزونه helm-diff

# نصب پکیج ردیس
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm install redis-release bitnami/redis \
  --set auth.enabled=false \
  --set architecture=standalone
🔒 اگر خواستی Redis رو با رمز عبور نصب کنی:
  --set auth.enabled=true --set auth.password=yourPassword

```

### ☸️ kubectl

```bash
# مشاهده وضعیت پادها
kubectl get pods

# مشاهده سرویس‌ها
kubectl get svc

# مشاهده لاگ یک پاد
kubectl logs <pod-name>

# مشاهده لاگ deployment
kubectl logs deployment/<deployment-name>

# پاک کردن یک پاد خاص (ریستارت خواهد شد)
kubectl delete pod <pod-name>

# port forward برای دسترسی محلی به پاد
kubectl port-forward deployment/<deployment-name> <local-port>:<container-port>

# مشاهده تمام منابع در namespace جاری
kubectl get all

# توصیف یک پاد برای جزئیات بیشتر
kubectl describe pod <pod-name>
```

### 🐳 Docker

```bash
# ساخت ایمیج داکر
docker build -t <image-name>:<tag> .

# مشاهده لیست ایمیج‌ها
docker images

# حذف ایمیج
docker rmi <image-id>
```

### 🟡 Minikube

```bash
# فعال‌سازی محیط داکر داخلی مینی‌کیوب برای build مستقیم
eval $(minikube docker-env)

# بارگذاری ایمیج ساخته شده به داخل مینی‌کیوب (در صورتی که از docker-env استفاده نشود)
minikube image load <image-name>:<tag>

# مشاهده IP مینی‌کیوب (برای سرویس نوع NodePort یا LoadBalancer)
minikube ip

# باز کردن داشبورد گرافیکی
minikube dashboard
```

---

### 🧹 دستورات مفید برای پاکسازی و رفع مشکلات

```bash
# پاکسازی کامل کش helm (در موارد مشکلات patch)
helm uninstall <release-name>

# حذف کامل یک پاد گیر کرده یا crash شده
kubectl delete pod <pod-name> --grace-period=0 --force

# بررسی خطاهای مربوط به probe
kubectl describe pod <pod-name>
```
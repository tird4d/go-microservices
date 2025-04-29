## 📦 Deployment Commands (Helm, kubectl, minikube)

### 🔁 Helm

```bash
# Create a new service/chart
helm create <release-name>

# Install or upgrade a release
helm upgrade --install <release-name> <chart-path>

# List all installed releases
helm list

# Uninstall a release
helm uninstall <release-name>

# Render final YAML output without applying it
helm template <release-name> <chart-path>

# Render and apply the upgrade
helm upgrade <release-name> <chart-path>

# Show rendered file differences before upgrade (requires helm-diff plugin)
helm diff upgrade <release-name> <chart-path>

# Add Redis chart repository and update
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

# Install Redis without authentication
helm install redis-release bitnami/redis \
  --set auth.enabled=false \
  --set architecture=standalone

helm install rabbitmq bitnami/rabbitmq  --namespace rabbitmq --create-namespace  --set auth.username=admin  --set auth.password=mypassword123   --set auth.erlangCookie=secretcookie123

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/prometheus



helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm install grafana grafana/grafana



helm install mongo-release bitnami/mongodb \
  --set auth.enabled=false \
  --set architecture=standalone


# Install Redis with password
# --set auth.enabled=true --set auth.password=yourPassword

# Render specific chart for review
helm template auth-service ./charts/auth-service
helm upgrade auth-service ./charts/auth-service

```

### ☸️ kubectl

```bash
# View the status of all pods
kubectl get pods

# View services in the current namespace
kubectl get svc

# View logs of a specific pod
kubectl logs <pod-name>

# View logs of a specific deployment
kubectl logs deployment/<deployment-name>

# Delete a specific pod (it will be restarted by the deployment)
kubectl delete pod <pod-name>

# Forward port from container to local machine
kubectl port-forward deployment/<deployment-name> <local-port>:<container-port>

# View all Kubernetes resources in the current namespace
kubectl get all

# Describe a pod in detail (for events, probe failures, etc.)
kubectl describe pod <pod-name>

# Get container names in a pod
kubectl get pod <pod-name> -o jsonpath="{.spec.containers[*].name}"

# Execute a command inside a pod interactively
kubectl exec -it <pod-name> -- /bin/sh

# Apply resources from a YAML file
kubectl apply -f <file.yaml>

# Delete resources from a YAML file
kubectl delete -f <file.yaml>

# Get architecture of cluster nodes (for image compatibility)
kubectl get node -o jsonpath="{.items[0].status.nodeInfo.architecture}"

# View recent cluster events (useful for probe/debugging issues)
kubectl get events --sort-by=.metadata.creationTimestamp

```

### 🟡 Minikube

```bash
# Enable local Docker environment for direct image build inside Minikube
eval $(minikube docker-env)

# Load an image manually into Minikube (when docker-env is not used)
minikube image load <image-name>:<tag>

# List images loaded into Minikube
minikube image list

# View Minikube IP address (for NodePort / LoadBalancer)
minikube ip

# Open the Kubernetes dashboard in browser
minikube dashboard

# SSH into the Minikube virtual machine
minikube ssh

minikube delete
minikube start

```


### Docker
```bash
# Build a Docker image
docker build -t <image-name>:<tag> .

# List all local Docker images
docker images

# Remove an image
docker rmi <image-id>

# Run a container interactively with shell access
docker run -it <image-name>:<tag> /bin/sh

```

### 🧹 Useful Cleanup & Debugging Commands


```bash
# Force uninstall a Helm release (used when patches or upgrades fail)
helm uninstall <release-name>

# Forcefully delete a stuck or crashing pod immediately
kubectl delete pod <pod-name> --grace-period=0 --force

# Describe pod to investigate health probe errors or crash reasons
kubectl describe pod <pod-name>

# Check if grpc-health-probe is present and executable
kubectl exec -it <pod-name> -- file /bin/grpc-health-probe

# Manually test a gRPC health check inside a pod
kubectl exec -it <pod-name> -- /bin/grpc-health-probe -addr=:50051
```
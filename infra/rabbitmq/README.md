# RabbitMQ — infra/rabbitmq

This folder documents and configures **RabbitMQ** as a cluster-level infrastructure dependency
deployed via the official **Bitnami Helm chart** (`bitnami/rabbitmq`).

> **This is a cluster-level infrastructure dependency, not an app chart.**  
> The Helm chart lives in the `bitnami` repo, not in `charts/`.  
> This folder holds the values override file and deployment instructions.  
> The CI/CD pipeline (`.github/workflows/rabbitmq-deploy.yml`) deploys it automatically on push.

---

## Why RabbitMQ

RabbitMQ is the **async message broker** used by three services:

```
order-service  ──publish──▶  [exchange: order.events]  ──route──▶  queue: email.notifications
                                                                          ▲
                                                           email-service  ┘  (consumes)

user-service   ──publish──▶  [exchange: user.events]   ──route──▶  queue: order.users
                                                                          ▲
                                                           order-service  ┘  (consumes)
```

| Service | Role | Queue / Exchange |
|---|---|---|
| `user-service` | Publisher | `user.events` exchange |
| `order-service` | Consumer + Publisher | Consumes `user.events`; publishes `order.events` |
| `email-service` | Consumer | Consumes `order.events` |

---

## Why Bitnami chart instead of a custom chart

| Custom chart (`charts/rabbitmq/`) | Bitnami chart (`bitnami/rabbitmq`) |
|---|---|
| Minimal — StatefulSet + Service only | Production-grade, 2000+ lines of config |
| No clustering support | Proper clustering with peer discovery |
| No TLS | TLS and mTLS support built-in |
| No PodDisruptionBudget | PDB, NetworkPolicy, RBAC included |
| No Prometheus metrics exporter | Metrics exporter sidecar built-in |
| We maintain it | Bitnami maintains it |

The approach: **Bitnami chart + our `values.yaml` overrides tracked in git** — identical philosophy
to how the monitoring stack uses `prometheus-community/kube-prometheus-stack`.

---

## What gets deployed

```
RabbitMQ (namespace: prod)
    ├── StatefulSet          — 1 replica (scale up for clustering)
    ├── Services             — ClusterIP for AMQP (5672) + management UI (15672)
    ├── Secret               — credentials (username/password/erlangCookie)
    ├── PersistentVolumeClaim — durable message storage
    └── ServiceMonitor       — Prometheus scrapes built-in metrics exporter
```

---

## Comparison: docker-compose vs EKS

| docker-compose | EKS (Bitnami chart) |
|---|---|
| `rabbitmq:3-management` image | `bitnami/rabbitmq:4.1.x` image |
| Credentials in `environment:` block | Credentials in Kubernetes Secret |
| No persistence (ephemeral) | PVC — messages survive pod restarts |
| Single node only | Configurable clustering |
| No metrics | Built-in Prometheus metrics exporter on `:9419` |
| Manual restart on failure | Kubernetes restarts the pod automatically |

---

## Connection strings

Services connect to RabbitMQ using the ClusterIP Service DNS name:

```
amqp://<username>:<password>@rabbitmq.prod.svc.cluster.local:5672/
```

Each service reads this from a Kubernetes Secret, injected via `envFrom` in its Helm chart:

| Service | Secret key |
|---|---|
| `auth-service` | `RABBITMQ_URI` |
| `order-service` | `RABBITMQ_URL` |
| `email-service` | `RABBITMQ_CONNECTION_STRING` |

---

## Deploy

### First install

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

helm upgrade --install rabbitmq bitnami/rabbitmq \
  --namespace prod \
  --create-namespace \
  -f infra/rabbitmq/values.yaml \
  --set auth.password=<strong-password> \
  --set auth.erlangCookie=<random-64-char-string> \
  --atomic \
  --timeout 5m
```

> Never put real credentials in `values.yaml` — pass them at deploy time via `--set`
> or via the CI/CD pipeline secrets (`RABBITMQ_PASSWORD`, `RABBITMQ_ERLANG_COOKIE`).

### Upgrade (chart or values change)

The CI/CD pipeline runs `helm upgrade --install` automatically when
`infra/rabbitmq/**` or `.github/workflows/rabbitmq-deploy.yml` changes on `main`.

### Access the management UI

```bash
kubectl port-forward svc/rabbitmq 15672:15672 -n prod
# Open: http://localhost:15672
# Username: value of auth.username in values.yaml
```

---

## Persistence

Messages are stored on a `PersistentVolumeClaim` backed by the cluster's default `StorageClass`
(EBS `gp3` on EKS). The PVC survives pod restarts and upgrades — messages in queues are **not lost**.

> To wipe all data (e.g. reset a test environment): delete the StatefulSet's PVC manually.

---

## Secrets management

Credentials are never stored in this file or in `values.yaml`.  
They live as GitHub Actions secrets and are passed to `helm upgrade` at deploy time:

| GitHub Actions Secret | Used for |
|---|---|
| `RABBITMQ_USERNAME` | RabbitMQ admin username |
| `RABBITMQ_PASSWORD` | RabbitMQ admin password |
| `RABBITMQ_ERLANG_COOKIE` | Cluster node authentication (must be same on all nodes) |

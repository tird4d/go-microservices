# 🛠 Microservice Testing & Quality Guide

This document defines the complete testing strategy for Go microservices deployed in Kubernetes environments.
It can be reused for every new microservice.

---

## 📋 1. Unit Tests

**Goal:**
- Validate internal functions and business logic independently.

**Tools:**
- `testing` package (standard Go)
- `testify` (for easier assertions)

**Example:**
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestValidateEmail(t *testing.T) {
    assert.True(t, ValidateEmail("test@example.com"))
    assert.False(t, ValidateEmail("invalidemail"))
}
```

**Command to run:**
```bash
go test ./...
```

---

## 📋 2. gRPC Functional Tests

**Goal:**
- Verify gRPC endpoints (like RegisterUser, LoginUser) are working as expected.

**Tools:**
- `grpcurl` (CLI tool)
- Postman (with gRPC support)
- Custom gRPC client scripts (optional)

**Example Test:**
```bash
grpcurl -plaintext localhost:50051 list

grpcurl -plaintext -d '{"email":"test@example.com","password":"123456"}' localhost:50051 user.UserService/RegisterUser
```

---

## 📋 3. Integration Tests

**Goal:**
- Ensure the service correctly interacts with dependencies like MongoDB, Redis, RabbitMQ.

**Tools:**
- Test containers in Kubernetes (already deployed)
- Fake/Mock servers (optional, for isolated testing)

**Example:**
- Insert a user and verify it persists in MongoDB.
- Publish a message to RabbitMQ and verify it is consumed correctly.

**Basic Plan:**
- Create a dedicated `integration_test.go`
- Connect to real services from test code (using Minikube/K8s services)

---

## 📋 4. Health Check Tests

**Goal:**
- Confirm the service is alive and healthy.

**Tools:**
- `grpc-health-probe` binary
- `curl` for HTTP metrics endpoint

**Example:**
```bash
# gRPC health check
grpc-health-probe -addr=localhost:50051

# HTTP metrics check
curl http://localhost:9090/metrics
```

**Expected Results:**
- `grpc-health-probe` returns `SERVING`
- `/metrics` endpoint responds with 200 OK

---

## 📋 5. Load Testing (Optional Advanced Stage)

**Goal:**
- Measure service performance under heavy traffic.

**Tools:**
- `ghz` (gRPC load testing tool)
- `k6` (HTTP load testing tool)

**Example:**
```bash
ghz --insecure --proto ./proto/user.proto --call user.UserService/RegisterUser -d '{"email":"test@example.com","password":"pass"}' -c 50 -n 1000 localhost:50051
```

**Metrics to track:**
- Latency
- Throughput
- Error rate

---

## 📋 6. CI/CD Integration (Later Stage)

**Goal:**
- Run all tests automatically during deployment.

**Tools:**
- GitHub Actions
- GitLab CI/CD
- Jenkins (optional)

**Example Steps:**
- Run `go test ./...` automatically on each pull request.
- Deploy to a staging namespace after successful test runs.

---

# 🎯 Summary Testing Flow

| Step | Type              | Priority |
|-----|-------------------|----------|
| 1   | Unit Tests         | Mandatory |
| 2   | gRPC Functional Tests | Mandatory |
| 3   | Integration Tests  | Highly Recommended |
| 4   | Health Checks      | Mandatory |
| 5   | Load Testing       | Optional (for production systems) |
| 6   | CI/CD Automation   | Recommended |

---

This guide should be followed for every new microservice created in your system 🚀

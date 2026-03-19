# 📊 Load Testing Guide with k6

Complete guide for load testing all microservices using k6 (Grafana's load testing tool).

## 🚀 Quick Start

```bash
# Run quick load test (30 seconds, 10 users)
cd load_tests
k6 run api_gateway_load_test.js --duration 30s --vus 10

# Run full test suite with all stages
bash run_load_tests.sh
```

## 📁 Test Scripts

### 1. **api_gateway_load_test.js**
Tests the API Gateway with realistic workload patterns:
- Product listing (most common)
- Single product lookup
- Category filtering
- User operations
- Health checks

**Key Metrics:**
- Throughput: 17+ requests/sec
- Average latency: ~300ms
- P95 latency: <500ms
- Error rate: <5%

**Stages:**
```
30s  : Ramp-up (0 → 100 users)
60s  : Sustained load (100 users)
10s  : Spike (100 → 200 users)
30s  : Recovery (200 → 100 users)
30s  : Stress test (100 → 300 users)
30s  : Cool-down (300 → 0 users)
```

### 2. **product_service_load_test.js**
Direct load test of Product Service endpoints:
- List all products
- Filter by category
- Get single product

**Load Profile:** 50-150 VUs across stages

### 3. **user_service_grpc_load_test.js**
gRPC load test for User Service:
- GetUser RPC
- ListUsers RPC

**Protocol:** gRPC over plaintext (port 50051)

## 📈 Metrics Explanation

### Standard HTTP Metrics
```
http_reqs              - Total requests sent
http_req_duration      - Time from request start to response end
http_req_failed        - Requests with 4xx/5xx status
http_req_blocked       - Time spent blocked
http_req_connecting    - TCP connection time
http_req_tls_handshaking - TLS handshake time
http_req_sending       - Request body sending time
http_req_waiting       - Server processing time
http_req_receiving     - Response body receiving time
```

### Custom Metrics
```
errors                 - Error rate (failed/total)
success                - Success rate (passed/total)
duration               - Request duration trend
rps                    - Requests per second counter
```

### Percentiles
```
p(50)  - Median (50% of requests below this)
p(95)  - 95th percentile (95% below this)
p(99)  - 99th percentile (99% below this)
```

## 🎯 Load Testing Scenarios

### 1. Light Load (Development/Testing)
```bash
k6 run api_gateway_load_test.js --vus 10 --duration 1m
```
- 10 concurrent users
- 1 minute duration
- Check basic functionality
- Verify no crashes

### 2. Medium Load (Production Baseline)
```bash
k6 run api_gateway_load_test.js --vus 50 --duration 5m
```
- 50 concurrent users
- 5 minutes duration
- ~850 requests/sec
- Check sustained performance

### 3. Heavy Load (Stress Test)
```bash
k6 run api_gateway_load_test.js --vus 200 --duration 5m
```
- 200 concurrent users
- 5 minutes duration
- ~3400 requests/sec
- Identify bottlenecks

### 4. Spike Test (Sudden Traffic)
```bash
k6 run -e SPIKE=true api_gateway_load_test.js --vus 50 --duration 5m
```
- Spike from 50 to 200 users in 10 seconds
- Measure recovery time
- Check autoscaling triggers

## 📊 Expected Results

### Healthy Service
```
✅ Error rate: < 1%
✅ p95 latency: < 300ms
✅ p99 latency: < 500ms
✅ Throughput: > 15 req/sec (per user)
✅ Success rate: > 99%
```

### Warning Signs
```
⚠️  Error rate: 1-5%
⚠️  p95 latency: 300-500ms
⚠️  p99 latency: 500-1000ms
⚠️  Throughput: 8-15 req/sec
⚠️  Success rate: 95-99%
```

### Critical Issues
```
❌ Error rate: > 5%
❌ p95 latency: > 500ms
❌ p99 latency: > 1000ms
❌ Throughput: < 8 req/sec
❌ Success rate: < 95%
```

## 🔍 Analyzing Results

### 1. View Results in Terminal
```bash
k6 run api_gateway_load_test.js
# Check final summary for key metrics
```

### 2. Export to JSON and Analyze
```bash
k6 run api_gateway_load_test.js --out json=results.json
# Import into analysis tool or process with jq
```

### 3. Real-time Dashboard (InfluxDB)
```bash
# Set up InfluxDB and Grafana (optional advanced setup)
k6 run api_gateway_load_test.js --out influxdb=http://localhost:8086
```

## 🐛 Troubleshooting

### Error: "Unable to establish connection"
```
✅ Solution: Verify services are running
docker compose ps
docker compose up -d api-gateway
```

### Error: "Threshold crossed"
```
✅ This is normal during stress tests
✅ Indicates performance degradation under load
✅ Review metrics to identify limits
```

### High Error Rate (>5%)
```
✅ Check service logs:
docker compose logs api-gateway
docker compose logs user-service

✅ Check resource usage:
docker stats

✅ Check for crashes:
docker compose ps
```

### High Latency (p95 > 500ms)
```
✅ Identify slow endpoint:
grep "duration.*<" test_output.log

✅ Check database performance:
docker exec go-micro-mongo mongostat

✅ Check network:
docker network inspect go-microservices_default
```

## 📋 Performance Optimization Checklist

If results show poor performance:

- [ ] Check database connection pooling
- [ ] Check API response payloads (reduce or paginate)
- [ ] Check for N+1 query problems
- [ ] Verify caching is working (Redis)
- [ ] Check for blocking operations
- [ ] Review gRPC call overhead
- [ ] Check for memory leaks
- [ ] Verify no blocking I/O in handlers
- [ ] Check error logs for exceptions
- [ ] Profile CPU usage

## 📈 Recommended Workflow

1. **Baseline Test (5 min)**
   ```bash
   k6 run api_gateway_load_test.js --vus 50 --duration 5m
   ```
   - Establish performance baseline
   - Document in project

2. **Spike Test (5 min)**
   ```bash
   k6 run -e SPIKE=true api_gateway_load_test.js
   ```
   - Test sudden traffic surge
   - Measure recovery

3. **Stress Test (until failure)**
   ```bash
   k6 run api_gateway_load_test.js --vus 300 --duration 10m
   ```
   - Find breaking point
   - Identify bottleneck

4. **Long-running Test (30 min)**
   ```bash
   k6 run api_gateway_load_test.js --vus 100 --duration 30m
   ```
   - Check for memory leaks
   - Verify stable performance

## 🎯 Performance Targets

| Metric | Target | Current |
|--------|--------|---------|
| Error Rate | < 1% | ? |
| p95 Latency | < 300ms | ? |
| p99 Latency | < 500ms | ? |
| Throughput | > 15 req/sec | ? |
| Success Rate | > 99% | ? |
| Max Users | > 100 | ? |

## 📚 References

- [k6 Documentation](https://k6.io/docs/)
- [k6 HTTP API](https://k6.io/docs/javascript-api/k6-http/)
- [k6 gRPC API](https://k6.io/docs/javascript-api/k6-net-grpc/)
- [Load Testing Best Practices](https://k6.io/docs/test-types/load-testing/)

---

**Next Steps:**
1. Run baseline test: `k6 run api_gateway_load_test.js --vus 50 --duration 5m`
2. Document baseline metrics
3. Optimize if needed
4. Run spike and stress tests
5. Update performance targets table above


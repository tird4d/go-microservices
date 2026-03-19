#!/bin/bash
# Load Testing Suite - Run all k6 tests and generate reports

set -e

WORKSPACE_ROOT="/home/tirdad/Projects/go-microservices"
LOAD_TESTS_DIR="$WORKSPACE_ROOT/load_tests"
RESULTS_DIR="$LOAD_TESTS_DIR/results"

# Create results directory
mkdir -p "$RESULTS_DIR"

echo "🚀 Microservices Load Testing Suite"
echo "===================================="
echo ""
echo "⏰ Started: $(date)"
echo "📁 Results directory: $RESULTS_DIR"
echo ""

# Function to run a test
run_test() {
  local test_name=$1
  local test_file=$2
  local vus=$3  # Virtual users
  local duration=$4
  
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "📊 Running: $test_name"
  echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo "VUs: $vus | Duration: $duration"
  echo ""
  
  k6 run "$test_file" \
    --out json="$RESULTS_DIR/${test_name}_results.json" \
    2>&1 | tee "$RESULTS_DIR/${test_name}_output.log"
  
  echo ""
  echo "✅ Completed: $test_name"
  echo ""
}

# Test 1: API Gateway Load Test
echo "🌐 PHASE 1: API Gateway Load Testing"
echo "===================================="
run_test "api-gateway" "$LOAD_TESTS_DIR/api_gateway_load_test.js"

sleep 5

# Test 2: Product Service Load Test  
echo "📦 PHASE 2: Product Service Load Testing"
echo "========================================"
run_test "product-service" "$LOAD_TESTS_DIR/product_service_load_test.js"

sleep 5

# Test 3: User Service gRPC Load Test
echo "👤 PHASE 3: User Service (gRPC) Load Testing"
echo "============================================"
run_test "user-service-grpc" "$LOAD_TESTS_DIR/user_service_grpc_load_test.js" || true

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📈 Load Testing Complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📁 Results saved to: $RESULTS_DIR"
echo "📋 Report files:"
ls -lh "$RESULTS_DIR"/ | grep -E "\.json|\.log" | awk '{print "   " $9 " (" $5 ")"}'
echo ""
echo "⏰ Finished: $(date)"
echo ""
echo "🔍 Next steps:"
echo "   1. Review results in $RESULTS_DIR"
echo "   2. Check error rates (should be < 5%)"
echo "   3. Check p95 latency (should be < 500ms)"
echo "   4. Identify bottlenecks"
echo ""

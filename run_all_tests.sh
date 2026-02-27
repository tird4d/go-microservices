#!/bin/bash
# Comprehensive test runner for all microservices
# Usage: bash run_all_tests.sh [--coverage] [--service SERVICE_NAME]

set -e

WORKSPACE_ROOT="/home/tirdad/Projects/go-microservices"
SHOW_COVERAGE=false
TARGET_SERVICE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage)
            SHOW_COVERAGE=true
            shift
            ;;
        --service)
            TARGET_SERVICE="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

cd "$WORKSPACE_ROOT"

if [ -n "$TARGET_SERVICE" ]; then
    # Run tests for specific service
    echo "🧪 Testing $TARGET_SERVICE..."
    cd "$TARGET_SERVICE"
    
    if [ "$SHOW_COVERAGE" = true ]; then
        echo "📊 With coverage analysis..."
        go test -v -cover ./...
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage_report.html
        echo "✅ Coverage report saved to $TARGET_SERVICE/coverage_report.html"
    else
        go test -v ./...
    fi
else
    # Run tests for all services
    services=("user_service" "auth_service" "product_service" "api_gateway")
    failed_services=()
    total_tests=0
    
    echo "🚀 Running Tests for All Microservices"
    echo "======================================="
    echo ""
    
    for service in "${services[@]}"; do
        if [ -d "$service" ]; then
            echo "📦 Testing: $service"
            cd "$service"
            
            if [ "$SHOW_COVERAGE" = true ]; then
                # Run with coverage
                output=$(go test -v -cover ./... 2>&1)
                echo "$output"
                coverage=$(echo "$output" | grep "coverage:" | tail -1 | grep -oP 'coverage: \K[0-9.]+' || echo "0")
                echo "  Coverage: ${coverage}%"
                
                # Generate HTML report
                go test -coverprofile=coverage.out ./...
                go tool cover -html=coverage.out -o coverage_report.html
                echo "  📄 Report: coverage_report.html"
            else
                if go test -v ./...; then
                    echo "  ✅ Passed"
                else
                    echo "  ❌ Failed"
                    failed_services+=("$service")
                fi
            fi
            
            cd ..
            echo ""
        fi
    done
    
    # Summary
    echo "======================================="
    echo "✨ Test Run Complete"
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        echo "✅ All services passed!"
    else
        echo "❌ Failed services: ${failed_services[*]}"
        exit 1
    fi
fi

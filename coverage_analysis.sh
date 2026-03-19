#!/bin/bash
# Coverage analysis tool for all microservices
# Generates HTML reports and displays coverage percentages

WORKSPACE_ROOT="/home/tirdad/Projects/go-microservices"
COVERAGE_TARGET=70

cd "$WORKSPACE_ROOT"

echo "📊 Coverage Analysis for All Microservices"
echo "=========================================="
echo ""

services=("user_service" "auth_service" "product_service" "api_gateway")
total_coverage=0
service_count=0

for service in "${services[@]}"; do
    if [ -d "$service" ]; then
        echo "Analyzing: $service"
        cd "$service"
        
        # Run tests and generate coverage
        go test -coverprofile=coverage.out ./... > /dev/null 2>&1
        
        # Extract coverage percentage
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        
        if [ -z "$coverage" ]; then
            coverage=0
        fi
        
        total_coverage=$(echo "$total_coverage + $coverage" | bc)
        service_count=$((service_count + 1))
        
        # Generate HTML report
        go tool cover -html=coverage.out -o coverage_report.html
        
        # Color coded output
        if (( $(echo "$coverage >= $COVERAGE_TARGET" | bc -l) )); then
            status="✅"
        else
            status="⚠️ "
        fi
        
        printf "%s %s: %.1f%%\n" "$status" "$service" "$coverage"
        
        cd ..
    fi
done

# Calculate average
if [ $service_count -gt 0 ]; then
    average=$(echo "scale=1; $total_coverage / $service_count" | bc)
    echo ""
    echo "=========================================="
    printf "📈 Average Coverage: %.1f%% (Target: %d%%)\n" "$average" "$COVERAGE_TARGET"
    
    if (( $(echo "$average >= $COVERAGE_TARGET" | bc -l) )); then
        echo "✅ Target met!"
    else
        echo "⚠️  Below target coverage"
    fi
fi

echo ""
echo "💾 HTML reports generated in each service directory:"
echo "   - Open service/coverage_report.html in browser to review"

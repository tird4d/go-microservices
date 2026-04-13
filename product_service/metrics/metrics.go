package metrics

import (
"github.com/prometheus/client_golang/prometheus"
)

var (
// Counter with endpoint label
RequestCounter = prometheus.NewCounterVec(
prometheus.CounterOpts{
Name: "product_service_requests_total",
Help: "Total number of requests to product-service, labeled by endpoint",
},
[]string{"endpoint"},
)

// Histogram for request duration
RequestDurationHistogram = prometheus.NewHistogramVec(
prometheus.HistogramOpts{
Name:    "product_service_request_duration_seconds",
Help:    "Histogram of response durations for product-service requests, labeled by endpoint",
Buckets: prometheus.DefBuckets,
},
[]string{"endpoint"},
)
)

// InitMetrics registers all metrics with the default global Prometheus registry
func InitMetrics() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDurationHistogram)
}

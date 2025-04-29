package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Counter with endpoint label
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_service_requests_total",
			Help: "Total number of requests to user-service, labeled by endpoint",
		},
		[]string{"endpoint"},
	)

	// Histogram for request duration
	RequestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_service_request_duration_seconds",
			Help:    "Histogram of response durations for user-service requests, labeled by endpoint",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

// Register all metrics
func InitMetrics() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDurationHistogram)
}

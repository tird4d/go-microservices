package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_service_requests_total",
			Help: "Total number of requests to order-service, labeled by endpoint",
		},
		[]string{"endpoint"},
	)

	RequestDurationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "order_service_request_duration_seconds",
			Help:    "Histogram of response durations for order-service requests, labeled by endpoint",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDurationHistogram)
}

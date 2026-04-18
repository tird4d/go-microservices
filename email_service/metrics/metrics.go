package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	EmailsSentCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "email_service_emails_total",
			Help: "Total number of emails processed by email-service, labeled by type",
		},
		[]string{"type"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(EmailsSentCounter)
}

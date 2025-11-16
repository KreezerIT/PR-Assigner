package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus метрики для сервиса
var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	ReviewerAssignmentsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reviewer_assignments_total",
			Help: "Total number of reviewer assignments",
		},
		[]string{"user_id"},
	)

	PRCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pull_requests_created_total",
			Help: "Total number of pull requests created",
		},
	)

	PRMergedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pull_requests_merged_total",
			Help: "Total number of pull requests merged",
		},
	)

	ReassignmentsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "reviewer_reassignments_total",
			Help: "Total number of reviewer reassignments",
		},
	)
)

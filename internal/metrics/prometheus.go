package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    HttpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"path", "method", "status"},
    )

    HttpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request latency",
            Buckets: prometheus.DefBuckets,
        },
        []string{"path", "method"},
    )

    AnomaliesTotal = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "anomalies_total",
            Help: "Total detected anomalies",
        },
    )
)

func Register() {
    prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration, AnomaliesTotal)
}

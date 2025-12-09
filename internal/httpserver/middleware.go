package httpserver

import (
    "net/http"
    "time"

    "github.com/lehater/go-k8s-analyzer/internal/metrics"
)

type statusRecorder struct {
    http.ResponseWriter
    status int
}

func (r *statusRecorder) WriteHeader(code int) {
    r.status = code
    r.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler, path string, method string) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rec := &statusRecorder{ResponseWriter: w, status: 200}
        next.ServeHTTP(rec, r)
        duration := time.Since(start).Seconds()
        metrics.HttpRequestsTotal.WithLabelValues(path, method, http.StatusText(rec.status)).Inc()
        metrics.HttpRequestDuration.WithLabelValues(path, method).Observe(duration)
    })
}

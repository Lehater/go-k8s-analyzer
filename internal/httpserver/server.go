package httpserver

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/example/go-k8s-analyzer/internal/analytics"
    "github.com/example/go-k8s-analyzer/internal/config"
    "github.com/example/go-k8s-analyzer/internal/logger"
    "github.com/example/go-k8s-analyzer/internal/metrics"
    "github.com/example/go-k8s-analyzer/internal/model"
    "github.com/example/go-k8s-analyzer/internal/storage"
)

type Server struct {
    httpServer *http.Server
    analyzer   *analytics.Analyzer
    redis      *storage.RedisStorage
    ingestCh   chan model.Metric
    log        *logger.Logger
}

func New(cfg config.Config, log *logger.Logger) *Server {
    an := analytics.NewAnalyzer(cfg.AnalyticsWindow)
    rs := storage.NewRedisStorage(cfg.RedisAddr, cfg.RedisDialTimeout, cfg.RedisReadTimeout, cfg.RedisWriteTimeout)
    metrics.Register()

    s := &Server{
        analyzer: an,
        redis:    rs,
        ingestCh: make(chan model.Metric, cfg.IngestBufferSize),
        log:      log,
    }

    mux := http.NewServeMux()
    mux.Handle("/ingest", metricsMiddleware(http.HandlerFunc(s.handleIngest), "/ingest", http.MethodPost))
    mux.Handle("/analyze", metricsMiddleware(http.HandlerFunc(s.handleAnalyze), "/analyze", http.MethodGet))
    mux.Handle("/metrics", promhttp.Handler())
    mux.HandleFunc("/healthz", s.handleHealth)

    s.httpServer = &http.Server{
        Addr:         cfg.HTTPAddr,
        Handler:      mux,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 5 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go s.runIngestLoop()

    return s
}

func (s *Server) runIngestLoop() {
    for m := range s.ingestCh {
        stats := s.analyzer.AddMetric(m)
        if stats.IsAnomaly {
            metrics.AnomaliesTotal.Inc()
        }
        ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
        _ = s.redis.SaveMetric(ctx, m.Timestamp.Format(time.RFC3339Nano), m, 10*time.Minute)
        cancel()
    }
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    defer r.Body.Close()

    var m model.Metric
    if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if m.Timestamp.IsZero() {
        m.Timestamp = time.Now()
    }
    if m.RPS < 0 {
        http.Error(w, "rps must be non-negative", http.StatusBadRequest)
        return
    }
    if m.CPU < 0 || m.CPU > 100 {
        http.Error(w, "cpu must be in [0,100]", http.StatusBadRequest)
        return
    }

    select {
    case s.ingestCh <- m:
        w.WriteHeader(http.StatusAccepted)
    default:
        http.Error(w, "ingest channel full", http.StatusServiceUnavailable)
    }
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    stats := s.analyzer.GetStats()
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}

func (s *Server) Start() error {
    s.log.Info("starting HTTP server on %s", s.httpServer.Addr)
    return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
    s.log.Info("stopping HTTP server")
    close(s.ingestCh)
    return s.httpServer.Shutdown(ctx)
}

package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/example/go-k8s-analyzer/internal/config"
    "github.com/example/go-k8s-analyzer/internal/httpserver"
    "github.com/example/go-k8s-analyzer/internal/logger"
    "github.com/example/go-k8s-analyzer/internal/model"
)

func newTestServer() *httptest.Server {
    cfg := config.Load()
    cfg.HTTPAddr = ":0"
    cfg.RedisAddr = "localhost:6379"
    logg := logger.New("debug")
    srv := httpserver.New(cfg, logg)
    return httptest.NewServer(srv.(*httpserver.Server).Handler())
}

// Simpler: just test handlers directly by constructing the mux is overkill;
// instead, call handle functions in more focused unit tests.
// For brevity, here basic integration with httptest.NewRecorder + handler.

func TestIngestAndAnalyze(t *testing.T) {
    cfg := config.Load()
    cfg.HTTPAddr = ":0"
    cfg.RedisAddr = "localhost:6379"
    logg := logger.New("debug")
    srv := httpserver.New(cfg, logg)

    // test ingest
    rec := httptest.NewRecorder()
    metric := model.Metric{
        Timestamp: time.Now(),
        CPU:       10,
        RPS:       100,
    }
    body, _ := json.Marshal(metric)
    req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))

    handler := http.HandlerFunc(srv.(*httpserver.Server).HandleIngestForTest) // if we exposed for tests
    handler.ServeHTTP(rec, req)
    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected 202, got %d", rec.Code)
    }
}

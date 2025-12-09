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

// Basic integration-style test for /ingest and /analyze handlers.
func TestIngestAndAnalyze(t *testing.T) {
    cfg := config.Load()
    cfg.HTTPAddr = ":0"
    cfg.RedisAddr = "localhost:6379"
    logg := logger.New("debug")
    srv := httpserver.New(cfg, logg)

    handler := srv.Handler()

    // send metric to /ingest
    metric := model.Metric{
        Timestamp: time.Now(),
        CPU:       10,
        RPS:       100,
    }
    body, _ := json.Marshal(metric)
    req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusAccepted {
        t.Fatalf("expected 202, got %d", rec.Code)
    }

    // give some time for ingest loop to process
    time.Sleep(10 * time.Millisecond)

    // query /analyze
    reqAnalyze := httptest.NewRequest(http.MethodGet, "/analyze", nil)
    recAnalyze := httptest.NewRecorder()
    handler.ServeHTTP(recAnalyze, reqAnalyze)

    if recAnalyze.Code != http.StatusOK {
        t.Fatalf("expected 200 from /analyze, got %d", recAnalyze.Code)
    }
}

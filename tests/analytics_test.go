package tests

import (
    "testing"

    "github.com/example/go-k8s-analyzer/internal/analytics"
    "github.com/example/go-k8s-analyzer/internal/model"
)

func TestAnalyzerEmpty(t *testing.T) {
    a := analytics.NewAnalyzer(50)
    stats := a.GetStats()
    if stats.Count != 0 {
        t.Fatalf("expected count 0, got %d", stats.Count)
    }
}

func TestAnalyzerWindowAndAnomaly(t *testing.T) {
    a := analytics.NewAnalyzer(5)
    for i := 0; i < 5; i++ {
        a.AddMetric(model.Metric{RPS: 100})
    }
    stats := a.GetStats()
    if stats.Count != 5 {
        t.Fatalf("expected count 5, got %d", stats.Count)
    }

    // add anomaly
    stats = a.AddMetric(model.Metric{RPS: 200})
    if !stats.IsAnomaly {
        t.Fatalf("expected anomaly for high RPS")
    }
}

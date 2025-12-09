package analytics

import (
    "math"
    "sync"

    "github.com/example/go-k8s-analyzer/internal/model"
)

type Stats struct {
    Count      int64   `json:"count"`
    Mean       float64 `json:"mean"`
    StdDev     float64 `json:"std_dev"`
    LastValue  float64 `json:"last_value"`
    ZScore     float64 `json:"z_score"`
    IsAnomaly  bool    `json:"is_anomaly"`
    WindowSize int     `json:"window_size"`
    AnomalyCnt int64   `json:"anomaly_count"`
}

type Analyzer struct {
    mu         sync.RWMutex
    windowSize int
    values     []float64
    sum        float64
    sumSquares float64
    anomalies  int64
}

func NewAnalyzer(windowSize int) *Analyzer {
    if windowSize <= 0 {
        windowSize = 50
    }
    return &Analyzer{
        windowSize: windowSize,
        values:     make([]float64, 0, windowSize),
    }
}

func (a *Analyzer) AddMetric(m model.Metric) Stats {
    a.mu.Lock()
    defer a.mu.Unlock()

    v := m.RPS
    if len(a.values) == a.windowSize {
        removed := a.values[0]
        a.values = a.values[1:]
        a.sum -= removed
        a.sumSquares -= removed * removed
    }
    a.values = append(a.values, v)
    a.sum += v
    a.sumSquares += v * v

    n := float64(len(a.values))
    mean := a.sum / n
    variance := (a.sumSquares / n) - (mean * mean)
    if variance < 0 {
        variance = 0
    }
    stdDev := math.Sqrt(variance)

    var z float64
    var isAnomaly bool
    if stdDev > 0 {
        z = math.Abs(v-mean) / stdDev
        isAnomaly = z > 2.0
        if isAnomaly {
            a.anomalies++
        }
    }

    return Stats{
        Count:      int64(len(a.values)),
        Mean:       mean,
        StdDev:     stdDev,
        LastValue:  v,
        ZScore:     z,
        IsAnomaly:  isAnomaly,
        WindowSize: a.windowSize,
        AnomalyCnt: a.anomalies,
    }
}

func (a *Analyzer) GetStats() Stats {
    a.mu.RLock()
    defer a.mu.RUnlock()

    if len(a.values) == 0 {
        return Stats{WindowSize: a.windowSize}
    }

    n := float64(len(a.values))
    mean := a.sum / n
    variance := (a.sumSquares / n) - (mean * mean)
    if variance < 0 {
        variance = 0
    }
    stdDev := math.Sqrt(variance)

    last := a.values[len(a.values)-1]
    var z float64
    var isAnomaly bool
    if stdDev > 0 {
        z = math.Abs(last-mean) / stdDev
        isAnomaly = z > 2.0
    }

    return Stats{
        Count:      int64(len(a.values)),
        Mean:       mean,
        StdDev:     stdDev,
        LastValue:  last,
        ZScore:     z,
        IsAnomaly:  isAnomaly,
        WindowSize: a.windowSize,
        AnomalyCnt: a.anomalies,
    }
}

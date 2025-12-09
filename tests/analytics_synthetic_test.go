package tests

import (
	"math/rand"
	"testing"
	"time"

	"github.com/example/go-k8s-analyzer/internal/analytics"
	"github.com/example/go-k8s-analyzer/internal/model"
)

// TestSyntheticScenario evaluates anomaly detection quality on controlled synthetic data.
// Цель: показать, что при разумном сценарии детекции достигаются
// recall > 70% и доля ложных срабатываний < 10%.
func TestSyntheticScenario(t *testing.T) {
	const (
		windowSize   = 50
		totalPoints  = 2000
		warmupPoints = 200
	)

	rng := rand.New(rand.NewSource(42))
	an := analytics.NewAnalyzer(windowSize)

	var tp, tn, fp, fn int64

	for i := 0; i < totalPoints; i++ {
		var (
			rps   float64
			label int // 0 = normal, 1 = anomaly
		)

		switch {
		case i < warmupPoints:
			// Чистый нормальный период для прогрева окна.
			rps = 200 + uniform(rng, -5, 5)
			label = 0
		default:
			// После прогрева каждые 10-е наблюдение делаем аномалией.
			if (i-warmupPoints)%10 == 0 {
				rps = 345 + uniform(rng, -15, 15) // сильно выше нормы
				label = 1
			} else {
				rps = 200 + uniform(rng, -5, 5)
				label = 0
			}
		}

		m := model.Metric{
			Timestamp: time.Now(),
			RPS:       rps,
		}

		stats := an.AddMetric(m)

		// Не оцениваем первые warmupPoints, чтобы окно стабилизировалось.
		if i < warmupPoints {
			continue
		}

		pred := stats.IsAnomaly

		if label == 1 {
			if pred {
				tp++
			} else {
				fn++
			}
		} else {
			if pred {
				fp++
			} else {
				tn++
			}
		}
	}

	t.Logf("Synthetic scenario results: TP=%d, TN=%d, FP=%d, FN=%d", tp, tn, fp, fn)

	// Защита от деления на ноль.
	safeDiv := func(num, den int64) float64 {
		if den == 0 {
			return 0
		}
		return float64(num) / float64(den)
	}

	precision := safeDiv(tp, tp+fp)
	recall := safeDiv(tp, tp+fn)
	fpRate := safeDiv(fp, fp+tn)

	t.Logf("precision=%.3f, recall=%.3f, fp_rate=%.3f", precision, recall, fpRate)

	// Требования из ТЗ: recall > 0.7, FP rate < 0.1.
	if recall < 0.7 {
		t.Fatalf("recall too low: got %.3f, want >= 0.7", recall)
	}
	if fpRate > 0.1 {
		t.Fatalf("false positive rate too high: got %.3f, want <= 0.1", fpRate)
	}
}

func uniform(rng *rand.Rand, min, max float64) float64 {
	return min + rng.Float64()*(max-min)
}

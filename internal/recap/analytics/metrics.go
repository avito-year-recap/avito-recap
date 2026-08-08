package analytics

import (
	"errors"
	"fmt"
	"time"

	"github.com/year-recap/internal/recap/model"
)

// EnrichMetrics calculates derived cohort-safe ratios. Callers that cross an
// external boundary should normalize metrics before invoking this function.
func EnrichMetrics(metrics model.Metrics) model.Metrics {
	metrics.RepeatRate = safeRate(metrics.RepeatedViews, metrics.TotalViews)
	metrics.PurchaseRate = safeRate(metrics.ChatsWithPurchase, metrics.ChatsStarted)
	return metrics
}

func safeRate(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

func SumUint64(values ...uint64) (uint64, error) {
	var result uint64
	for _, value := range values {
		if value > ^uint64(0)-result {
			return 0, errors.New("event counters overflow uint64")
		}
		result += value
	}
	return result, nil
}

var (
	ErrInvalidYear     = errors.New("invalid recap year")
	ErrYearNotComplete = errors.New("recap year is not complete")
)

func CompletedYearPeriod(year uint32, now time.Time) (model.RecapPeriod, error) {
	now = now.UTC()
	if year == 0 || year > uint32(now.Year()) {
		return model.RecapPeriod{}, ErrInvalidYear
	}
	if year == uint32(now.Year()) {
		return model.RecapPeriod{}, fmt.Errorf("%w: %d is still in progress", ErrYearNotComplete, year)
	}
	start := time.Date(int(year), time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(int(year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	return model.RecapPeriod{Year: year, StartAt: start, EndAt: end, Final: true}, nil
}

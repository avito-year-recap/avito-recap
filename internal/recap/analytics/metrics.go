package analytics

import "github.com/year-recap/internal/recap/model"

// EnrichMetrics normalizes strings and recalculates cohort-safe ratios.
func EnrichMetrics(metrics model.Metrics) model.Metrics {
	metrics = model.NormalizeMetrics(metrics)
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

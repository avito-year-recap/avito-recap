package recap

// EnrichMetrics normalizes strings and calculates only ratios whose numerator is
// an explicitly linked subset of the denominator. The ratios are always
// recomputed so rule evaluation cannot depend on stale input.
func EnrichMetrics(metrics Metrics) Metrics {
	metrics = normalizeMetrics(metrics)
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

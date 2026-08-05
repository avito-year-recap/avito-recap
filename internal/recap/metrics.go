package recap

// EnrichMetrics calculates derived ratios from validated base counters.
// The ratios are always recomputed so rule evaluation cannot depend on stale input.
func EnrichMetrics(metrics Metrics) Metrics {
	metrics.FavoriteRate = safeRate(metrics.FavoritesAdded, metrics.TotalViews)
	metrics.ChatRate = safeRate(metrics.ChatsStarted, metrics.TotalViews)
	metrics.RepeatRate = safeRate(metrics.RepeatedViews, metrics.TotalViews)
	metrics.PublicationRate = safeRate(metrics.ListingsPublished, metrics.ListingsCreated)
	metrics.SaleRate = safeRate(metrics.SalesCompleted, metrics.ListingsPublished)
	metrics.PurchaseRate = safeRate(metrics.ChatsWithPurchase, metrics.ChatsStarted)

	return metrics
}

func safeRate(part, total uint64) float64 {
	if total == 0 {
		return 0
	}

	return float64(part) / float64(total)
}

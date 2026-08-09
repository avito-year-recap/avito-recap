package structural

import (
	"fmt"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"math"
	"time"
)

func ValidateMetrics(metrics model.Metrics) error {
	metrics = model.NormalizeMetrics(metrics)
	knownEvents, err := analytics.SumUint64(metrics.Searches, metrics.TotalViews, metrics.FavoritesAdded,
		metrics.ChatsStarted, metrics.ListingsCreated, metrics.ListingsPublished,
		metrics.PurchasesCompleted, metrics.SalesCompleted)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidMetrics, err)
	}
	if knownEvents > metrics.TotalEvents {
		return fmt.Errorf("%w: known event counters (%d) exceed total events (%d)", ErrInvalidMetrics, knownEvents, metrics.TotalEvents)
	}
	if metrics.UniqueListings > metrics.TotalViews {
		return fmt.Errorf("%w: unique listings exceed total views", ErrInvalidMetrics)
	}
	if metrics.RepeatedViews != metrics.TotalViews-metrics.UniqueListings {
		return fmt.Errorf("%w: repeated views must equal total views minus unique listings", ErrInvalidMetrics)
	}
	if metrics.ChatsWithPurchase > metrics.ChatsStarted {
		return fmt.Errorf("%w: chats with purchase exceed started chats", ErrInvalidMetrics)
	}
	if metrics.TopCategoryViews > metrics.TotalViews {
		return fmt.Errorf("%w: top-category views exceed total views", ErrInvalidMetrics)
	}
	if (metrics.TopCategoryCode == "") != (metrics.TopCategory == "") {
		return fmt.Errorf("%w: top category code and title must be set together", ErrInvalidMetrics)
	}
	if metrics.TopCategory == "" && metrics.TopCategoryViews != 0 {
		return fmt.Errorf("%w: top category is empty but its view count is non-zero", ErrInvalidMetrics)
	}
	if metrics.TopCategory != "" && metrics.TopCategoryViews == 0 {
		return fmt.Errorf("%w: top category is set but its view count is zero", ErrInvalidMetrics)
	}
	if metrics.TopCategoryShareable && metrics.TopCategory == "" {
		return fmt.Errorf("%w: empty top category cannot be shareable", ErrInvalidMetrics)
	}
	if metrics.CategoriesCount > metrics.TotalEvents {
		return fmt.Errorf("%w: category count exceeds total events", ErrInvalidMetrics)
	}
	if metrics.ActiveDays > metrics.TotalEvents {
		return fmt.Errorf("%w: active days exceed total events", ErrInvalidMetrics)
	}
	if metrics.TopCategoryCode != "" && !isSafeCategoryCode(metrics.TopCategoryCode) {
		return fmt.Errorf("%w: top category code is unsafe", ErrInvalidMetrics)
	}
	if metrics.MostActiveMonth > 12 {
		return fmt.Errorf("%w: active month must be in range 0..12", ErrInvalidMetrics)
	}
	seenCategoryActivities := make(map[string]struct{}, len(metrics.CategoryActivities))
	var categoryViews, categoryFavorites, categoryPurchases uint64
	for index, activity := range metrics.CategoryActivities {
		if activity.CategoryCode == "" || activity.Category == "" {
			return fmt.Errorf("%w: category activity %d requires code and title", ErrInvalidMetrics, index)
		}
		if !isSafeCategoryCode(activity.CategoryCode) {
			return fmt.Errorf("%w: category activity %d has unsafe code", ErrInvalidMetrics, index)
		}
		if _, exists := seenCategoryActivities[activity.CategoryCode]; exists {
			return fmt.Errorf("%w: duplicate category activity %q", ErrInvalidMetrics, activity.CategoryCode)
		}
		seenCategoryActivities[activity.CategoryCode] = struct{}{}
		if activity.Views == 0 && activity.FavoritesAdded == 0 && activity.PurchasesCompleted == 0 {
			return fmt.Errorf("%w: category activity %q has no evidence", ErrInvalidMetrics, activity.CategoryCode)
		}
		var err error
		categoryViews, err = analytics.SumUint64(categoryViews, activity.Views)
		if err != nil {
			return fmt.Errorf("%w: category views overflow", ErrInvalidMetrics)
		}
		categoryFavorites, err = analytics.SumUint64(categoryFavorites, activity.FavoritesAdded)
		if err != nil {
			return fmt.Errorf("%w: category favorites overflow", ErrInvalidMetrics)
		}
		categoryPurchases, err = analytics.SumUint64(categoryPurchases, activity.PurchasesCompleted)
		if err != nil {
			return fmt.Errorf("%w: category purchases overflow", ErrInvalidMetrics)
		}
	}
	if uint64(len(metrics.CategoryActivities)) > metrics.CategoriesCount {
		return fmt.Errorf("%w: detailed category count exceeds categories count", ErrInvalidMetrics)
	}
	if categoryViews > metrics.TotalViews || categoryFavorites > metrics.FavoritesAdded || categoryPurchases > metrics.PurchasesCompleted {
		return fmt.Errorf("%w: per-category counters exceed aggregate counters", ErrInvalidMetrics)
	}
	return nil
}

func ValidateMetricsForPeriod(metrics model.Metrics, period model.RecapPeriod) error {
	if err := ValidateMetrics(metrics); err != nil {
		return err
	}
	maxDays := uint64(period.EndAt.Sub(period.StartAt) / (24 * time.Hour))
	if maxDays == 0 || metrics.ActiveDays > maxDays {
		return fmt.Errorf("%w: active days %d exceed period length %d", ErrInvalidMetrics, metrics.ActiveDays, maxDays)
	}
	if metrics.TotalEvents > 0 && (metrics.MostActiveMonth < 1 || metrics.MostActiveMonth > 12) {
		return fmt.Errorf("%w: active month is required for a non-empty annual period", ErrInvalidMetrics)
	}
	return nil
}

func ValidateStoredRates(metrics model.Metrics) error {
	expected := analytics.EnrichMetrics(metrics)
	checks := []struct {
		name             string
		actual, expected float64
	}{
		{name: "repeat rate", actual: metrics.RepeatRate, expected: expected.RepeatRate},
		{name: "purchase rate", actual: metrics.PurchaseRate, expected: expected.PurchaseRate},
	}
	for _, check := range checks {
		if math.IsNaN(check.actual) || math.IsInf(check.actual, 0) || math.Abs(check.actual-check.expected) > 1e-12 {
			return fmt.Errorf("%s is inconsistent: got %v, want %v", check.name, check.actual, check.expected)
		}
	}
	return nil
}

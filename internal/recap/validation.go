package recap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
	ErrInvalidMetrics = errors.New("invalid metrics")
)

func validateProfile(profile Profile) error {
	if strings.TrimSpace(profile.ID) == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidProfile)
	}
	if strings.TrimSpace(profile.DisplayName) == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidProfile)
	}

	return nil
}

func isUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}

	compact := value[:8] + value[9:13] + value[14:18] + value[19:23] + value[24:]
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != 16 {
		return false
	}

	for _, value := range decoded {
		if value != 0 {
			return true
		}
	}

	return false
}

func validateMetrics(metrics Metrics) error {
	knownEvents := uint64(0)
	for _, count := range []uint64{
		metrics.TotalViews,
		metrics.FavoritesAdded,
		metrics.ChatsStarted,
		metrics.ListingsCreated,
		metrics.ListingsPublished,
		metrics.PurchasesCompleted,
		metrics.SalesCompleted,
	} {
		if count > ^uint64(0)-knownEvents {
			return fmt.Errorf("%w: known event counters overflow uint64", ErrInvalidMetrics)
		}
		knownEvents += count
	}

	if knownEvents > metrics.TotalEvents {
		return fmt.Errorf(
			"%w: known event counters (%d) exceed total events (%d)",
			ErrInvalidMetrics,
			knownEvents,
			metrics.TotalEvents,
		)
	}
	if metrics.UniqueListings > metrics.TotalViews {
		return fmt.Errorf("%w: unique listings exceed total views", ErrInvalidMetrics)
	}
	if metrics.RepeatedViews > metrics.TotalViews {
		return fmt.Errorf("%w: repeated views exceed total views", ErrInvalidMetrics)
	}
	if metrics.TopCategoryViews > metrics.TotalViews {
		return fmt.Errorf("%w: top-category views exceed total views", ErrInvalidMetrics)
	}
	if strings.TrimSpace(metrics.TopCategory) == "" && metrics.TopCategoryViews != 0 {
		return fmt.Errorf("%w: top category is empty but its view count is non-zero", ErrInvalidMetrics)
	}
	if strings.TrimSpace(metrics.TopCategory) != "" && metrics.TopCategoryViews == 0 {
		return fmt.Errorf("%w: top category is set but its view count is zero", ErrInvalidMetrics)
	}
	if metrics.TopCategoryShareable && strings.TrimSpace(metrics.TopCategory) == "" {
		return fmt.Errorf("%w: empty top category cannot be shareable", ErrInvalidMetrics)
	}
	if metrics.CategoriesCount > metrics.TotalEvents {
		return fmt.Errorf("%w: category count exceeds total events", ErrInvalidMetrics)
	}
	if metrics.ActiveDays > metrics.TotalEvents && metrics.TotalEvents > 0 {
		return fmt.Errorf("%w: active days exceed total events", ErrInvalidMetrics)
	}
	if metrics.MostActiveMonth > 12 {
		return fmt.Errorf("%w: active month must be in range 0..12", ErrInvalidMetrics)
	}
	if metrics.ActiveDays > 366 {
		return fmt.Errorf("%w: active days cannot exceed 366", ErrInvalidMetrics)
	}

	return nil
}

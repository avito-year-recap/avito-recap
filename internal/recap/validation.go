package recap

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
	ErrInvalidMetrics = errors.New("invalid metrics")
)

func validateProfile(profile Profile) error {
	if profile.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", ErrInvalidProfile)
	}
	if strings.TrimSpace(profile.Code) == "" {
		return fmt.Errorf("%w: code is required", ErrInvalidProfile)
	}
	if strings.TrimSpace(profile.DisplayName) == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidProfile)
	}

	return nil
}

func validateMetrics(metrics Metrics) error {
	knownEvents, err := sumUint64(
		metrics.Searches,
		metrics.TotalViews,
		metrics.FavoritesAdded,
		metrics.ChatsStarted,
		metrics.ListingsCreated,
		metrics.ListingsPublished,
		metrics.PurchasesCompleted,
		metrics.SalesCompleted,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMetrics, err)
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
	expectedRepeated := metrics.TotalViews - metrics.UniqueListings

	if metrics.RepeatedViews != expectedRepeated {
		return fmt.Errorf(
			"%w: repeated views must equal total views minus unique listings",
			ErrInvalidMetrics,
		)
	}

	if metrics.RepeatedViews > metrics.TotalViews {
		return fmt.Errorf("%w: repeated views exceed total views", ErrInvalidMetrics)
	}
	//// Counters for different action types are not a closed annual funnel
	if metrics.ChatsWithPurchase > metrics.ChatsStarted {
		return fmt.Errorf("%w: chats with purchase exceed started chats", ErrInvalidMetrics)
	}

	if metrics.TopCategoryViews > metrics.TotalViews {
		return fmt.Errorf("%w: top-category views exceed total views", ErrInvalidMetrics)
	}

	categoryCode := strings.TrimSpace(metrics.TopCategoryCode)
	categoryTitle := strings.TrimSpace(metrics.TopCategory)
	if (categoryCode == "") != (categoryTitle == "") {
		return fmt.Errorf("%w: top category code and title must be set together", ErrInvalidMetrics)
	}
	if categoryTitle == "" && metrics.TopCategoryViews != 0 {
		return fmt.Errorf("%w: top category is empty but its view count is non-zero", ErrInvalidMetrics)
	}
	if categoryTitle != "" && metrics.TopCategoryViews == 0 {
		return fmt.Errorf("%w: top category is set but its view count is zero", ErrInvalidMetrics)
	}
	if metrics.TopCategoryShareable && categoryTitle == "" {
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

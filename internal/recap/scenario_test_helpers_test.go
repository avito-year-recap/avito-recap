package recap

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

var ErrInvalidScenario = errors.New("invalid seed scenario")

type SeedScenario struct {
	ProfileCode string             `json:"profileCode"`
	Year        uint32             `json:"year"`
	RandomSeed  int64              `json:"randomSeed"`
	Activity    SeedActivity       `json:"activity"`
	Categories  []WeightedCategory `json:"categories"`
	Months      []WeightedMonth    `json:"months"`
}

type SeedActivity struct {
	Searches           uint64 `json:"searches"`
	ListingViews       uint64 `json:"listingViews"`
	UniqueListings     uint64 `json:"uniqueListings"`
	FavoritesAdded     uint64 `json:"favoritesAdded"`
	ChatsStarted       uint64 `json:"chatsStarted"`
	ChatsWithPurchase  uint64 `json:"chatsWithPurchase"`
	ListingsCreated    uint64 `json:"listingsCreated"`
	ListingsPublished  uint64 `json:"listingsPublished"`
	PurchasesCompleted uint64 `json:"purchasesCompleted"`
	SalesCompleted     uint64 `json:"salesCompleted"`
	ActiveDays         uint64 `json:"activeDays"`
}
type WeightedCategory struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Weight    uint32 `json:"weight"`
	Shareable bool   `json:"shareable"`
}

type WeightedMonth struct {
	Month  uint32 `json:"month"`
	Weight uint32 `json:"weight"`
}

// MetricsFromScenario turns aggregate seed activity into the same metrics contract
// that a production analytics storage would return. Ties are resolved by code/month,
// so the result stays reproducible even if JSON item order changes.
func MetricsFromScenario(scenario SeedScenario) (Metrics, error) {
	if strings.TrimSpace(scenario.ProfileCode) == "" {
		return Metrics{}, fmt.Errorf("%w: profile code is required", ErrInvalidScenario)
	}
	if scenario.Year == 0 {
		return Metrics{}, fmt.Errorf("%w: year is required", ErrInvalidScenario)
	}
	if scenario.Activity.UniqueListings > scenario.Activity.ListingViews {
		return Metrics{}, fmt.Errorf("%w: unique listings exceed views", ErrInvalidScenario)
	}

	topCategory, err := pickTopCategory(scenario.Categories)
	if err != nil {
		return Metrics{}, err
	}
	activeMonth, err := pickActiveMonth(scenario.Months)
	if err != nil {
		return Metrics{}, err
	}

	totalEvents, err := sumUint64(
		scenario.Activity.Searches,
		scenario.Activity.ListingViews,
		scenario.Activity.FavoritesAdded,
		scenario.Activity.ChatsStarted,
		scenario.Activity.ListingsCreated,
		scenario.Activity.ListingsPublished,
		scenario.Activity.PurchasesCompleted,
		scenario.Activity.SalesCompleted,
	)
	if err != nil {
		return Metrics{}, fmt.Errorf("%w: %v", ErrInvalidScenario, err)
	}
	if scenario.Activity.ChatsWithPurchase > scenario.Activity.ChatsStarted {
		return Metrics{}, fmt.Errorf(
			"%w: chats with purchase exceed started chats",
			ErrInvalidScenario,
		)
	}

	metrics := Metrics{
		TotalEvents:          totalEvents,
		Searches:             scenario.Activity.Searches,
		TotalViews:           scenario.Activity.ListingViews,
		UniqueListings:       scenario.Activity.UniqueListings,
		RepeatedViews:        scenario.Activity.ListingViews - scenario.Activity.UniqueListings,
		FavoritesAdded:       scenario.Activity.FavoritesAdded,
		ChatsStarted:         scenario.Activity.ChatsStarted,
		ChatsWithPurchase:    scenario.Activity.ChatsWithPurchase,
		ListingsCreated:      scenario.Activity.ListingsCreated,
		ListingsPublished:    scenario.Activity.ListingsPublished,
		PurchasesCompleted:   scenario.Activity.PurchasesCompleted,
		SalesCompleted:       scenario.Activity.SalesCompleted,
		ActiveDays:           scenario.Activity.ActiveDays,
		CategoriesCount:      uint64(len(scenario.Categories)),
		TopCategoryCode:      topCategory.Code,
		TopCategory:          topCategory.Title,
		TopCategoryViews:     weightedCount(scenario.Activity.ListingViews, topCategory.Weight),
		TopCategoryShareable: topCategory.Shareable,
		MostActiveMonth:      activeMonth.Month,
	}

	if err := validateMetrics(metrics); err != nil {
		return Metrics{}, fmt.Errorf("%w: %v", ErrInvalidScenario, err)
	}

	return EnrichMetrics(metrics), nil
}

func pickTopCategory(categories []WeightedCategory) (WeightedCategory, error) {
	if len(categories) == 0 {
		return WeightedCategory{}, fmt.Errorf("%w: categories are required", ErrInvalidScenario)
	}

	items := append([]WeightedCategory(nil), categories...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Weight != items[j].Weight {
			return items[i].Weight > items[j].Weight
		}
		return items[i].Code < items[j].Code
	})

	seen := make(map[string]struct{}, len(items))
	weightSum := uint64(0)
	for _, category := range items {
		code := strings.TrimSpace(category.Code)
		title := strings.TrimSpace(category.Title)
		if code == "" || title == "" {
			return WeightedCategory{}, fmt.Errorf("%w: category code and title are required", ErrInvalidScenario)
		}
		if category.Weight == 0 {
			return WeightedCategory{}, fmt.Errorf("%w: category %q has zero weight", ErrInvalidScenario, code)
		}
		if _, exists := seen[code]; exists {
			return WeightedCategory{}, fmt.Errorf("%w: duplicate category %q", ErrInvalidScenario, code)
		}
		seen[code] = struct{}{}
		weightSum += uint64(category.Weight)
	}
	if weightSum != 100 {
		return WeightedCategory{}, fmt.Errorf("%w: category weights sum to %d, want 100", ErrInvalidScenario, weightSum)
	}

	return items[0], nil
}

func pickActiveMonth(months []WeightedMonth) (WeightedMonth, error) {
	if len(months) == 0 {
		return WeightedMonth{}, fmt.Errorf("%w: months are required", ErrInvalidScenario)
	}

	items := append([]WeightedMonth(nil), months...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Weight != items[j].Weight {
			return items[i].Weight > items[j].Weight
		}
		return items[i].Month < items[j].Month
	})

	seen := make(map[uint32]struct{}, len(items))
	weightSum := uint64(0)
	for _, month := range items {
		if month.Month < 1 || month.Month > 12 {
			return WeightedMonth{}, fmt.Errorf("%w: month %d is outside 1..12", ErrInvalidScenario, month.Month)
		}
		if month.Weight == 0 {
			return WeightedMonth{}, fmt.Errorf("%w: month %d has zero weight", ErrInvalidScenario, month.Month)
		}
		if _, exists := seen[month.Month]; exists {
			return WeightedMonth{}, fmt.Errorf("%w: duplicate month %d", ErrInvalidScenario, month.Month)
		}
		seen[month.Month] = struct{}{}
		weightSum += uint64(month.Weight)
	}
	if weightSum != 100 {
		return WeightedMonth{}, fmt.Errorf("%w: month weights sum to %d, want 100", ErrInvalidScenario, weightSum)
	}

	return items[0], nil
}

func weightedCount(total uint64, weight uint32) uint64 {
	if total == 0 || weight == 0 {
		return 0
	}

	value := math.Round(float64(total) * float64(weight) / 100)
	if value < 1 {
		return 1
	}
	return uint64(value)
}

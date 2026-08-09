package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/validation"
)

type SeedStorage interface {
	UpsertProfiles(context.Context, []model.Profile) error
	UpsertAnnualMetrics(context.Context, uuid.UUID, uint32, model.Metrics) error
	PutActionableState(context.Context, uuid.UUID, time.Time, model.ActionableState) error
}

type scenario struct {
	ProfileCode     string                `json:"profileCode"`
	Year            uint32                `json:"year"`
	Activity        activity              `json:"activity"`
	Categories      []weightedCategory    `json:"categories"`
	Months          []weightedMonth       `json:"months"`
	ActionableState model.ActionableState `json:"actionableState"`
}

type activity struct {
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

type weightedCategory struct {
	Code               string `json:"code"`
	Title              string `json:"title"`
	Weight             uint32 `json:"weight"`
	Shareable          bool   `json:"shareable"`
	Views              uint64 `json:"views,omitempty"`
	FavoritesAdded     uint64 `json:"favoritesAdded,omitempty"`
	PurchasesCompleted uint64 `json:"purchasesCompleted,omitempty"`
}

type weightedMonth struct {
	Month  uint32 `json:"month"`
	Weight uint32 `json:"weight"`
}

func LoadDemoData(ctx context.Context, storage SeedStorage, profilesPath, scenariosPath string) error {
	profiles, err := readJSON[[]model.Profile](profilesPath)
	if err != nil {
		return fmt.Errorf("load profiles seed: %w", err)
	}
	scenarios, err := readJSON[[]scenario](scenariosPath)
	if err != nil {
		return fmt.Errorf("load scenarios seed: %w", err)
	}
	byCode := make(map[string]model.Profile, len(profiles))
	for index := range profiles {
		profiles[index] = model.NormalizeProfile(profiles[index])
		profile := profiles[index]
		if err := validation.ValidateProfile(profile); err != nil {
			return fmt.Errorf("invalid profile seed at index %d: %w", index, err)
		}
		if _, exists := byCode[profile.Code]; exists {
			return fmt.Errorf("duplicate profile code %q", profile.Code)
		}
		byCode[profile.Code] = profile
	}
	if err := storage.UpsertProfiles(ctx, profiles); err != nil {
		return fmt.Errorf("seed profiles: %w", err)
	}

	observedAt := time.Now().UTC()
	for _, item := range scenarios {
		profile, ok := byCode[strings.TrimSpace(item.ProfileCode)]
		if !ok {
			return fmt.Errorf("scenario references unknown profile code %q", item.ProfileCode)
		}
		metrics, err := metricsFromScenario(item)
		if err != nil {
			return fmt.Errorf("scenario %s/%d: %w", profile.Code, item.Year, err)
		}
		if err := storage.UpsertAnnualMetrics(ctx, profile.ID, item.Year, metrics); err != nil {
			return fmt.Errorf("seed annual metrics for %s/%d: %w", profile.Code, item.Year, err)
		}
		if err := storage.PutActionableState(ctx, profile.ID, observedAt, item.ActionableState); err != nil {
			return fmt.Errorf("seed actionable state for %s: %w", profile.Code, err)
		}
	}
	return nil
}

func readJSON[T any](path string) (T, error) {
	var zero T
	data, err := os.ReadFile(path)
	if err != nil {
		return zero, err
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return zero, err
	}
	return value, nil
}

func metricsFromScenario(item scenario) (model.Metrics, error) {
	if item.Year == 0 || strings.TrimSpace(item.ProfileCode) == "" {
		return model.Metrics{}, fmt.Errorf("profileCode and year are required")
	}
	if item.Activity.UniqueListings > item.Activity.ListingViews {
		return model.Metrics{}, fmt.Errorf("unique listings exceed listing views")
	}
	if item.Activity.ChatsWithPurchase > item.Activity.ChatsStarted {
		return model.Metrics{}, fmt.Errorf("chats with purchase exceed chats started")
	}

	top, err := selectTopCategory(item.Categories)
	if err != nil {
		return model.Metrics{}, err
	}
	month, err := selectTopMonth(item.Months)
	if err != nil {
		return model.Metrics{}, err
	}
	total, err := analytics.SumUint64(
		item.Activity.Searches,
		item.Activity.ListingViews,
		item.Activity.FavoritesAdded,
		item.Activity.ChatsStarted,
		item.Activity.ListingsCreated,
		item.Activity.ListingsPublished,
		item.Activity.PurchasesCompleted,
		item.Activity.SalesCompleted,
	)
	if err != nil {
		return model.Metrics{}, err
	}

	categoryActivities := make([]model.CategoryActivity, 0, len(item.Categories))
	for _, category := range item.Categories {
		if category.Views == 0 && category.FavoritesAdded == 0 && category.PurchasesCompleted == 0 {
			continue
		}
		categoryActivities = append(categoryActivities, model.CategoryActivity{
			CategoryCode:       strings.TrimSpace(category.Code),
			Category:           strings.TrimSpace(category.Title),
			Shareable:          category.Shareable,
			Views:              category.Views,
			FavoritesAdded:     category.FavoritesAdded,
			PurchasesCompleted: category.PurchasesCompleted,
		})
	}

	metrics := model.Metrics{
		TotalEvents:        total,
		Searches:           item.Activity.Searches,
		TotalViews:         item.Activity.ListingViews,
		UniqueListings:     item.Activity.UniqueListings,
		RepeatedViews:      item.Activity.ListingViews - item.Activity.UniqueListings,
		FavoritesAdded:     item.Activity.FavoritesAdded,
		ChatsStarted:       item.Activity.ChatsStarted,
		ChatsWithPurchase:  item.Activity.ChatsWithPurchase,
		ListingsCreated:    item.Activity.ListingsCreated,
		ListingsPublished:  item.Activity.ListingsPublished,
		PurchasesCompleted: item.Activity.PurchasesCompleted,
		SalesCompleted:     item.Activity.SalesCompleted,
		ActiveDays:         item.Activity.ActiveDays,
		CategoriesCount:    uint64(len(item.Categories)),
		MostActiveMonth:    month.Month,
		CategoryActivities: categoryActivities,
	}
	if top.Code != "" {
		metrics.TopCategoryCode = top.Code
		metrics.TopCategory = top.Title
		metrics.TopCategoryShareable = top.Shareable
		metrics.TopCategoryViews = weightedCount(item.Activity.ListingViews, top.Weight)
	}
	metrics = analytics.EnrichMetrics(model.NormalizeMetrics(metrics))
	if err := validation.ValidateMetrics(metrics); err != nil {
		return model.Metrics{}, err
	}
	return metrics, nil
}

func selectTopCategory(values []weightedCategory) (weightedCategory, error) {
	if len(values) == 0 {
		return weightedCategory{}, nil
	}
	values = append([]weightedCategory(nil), values...)
	sort.Slice(values, func(i, j int) bool {
		if values[i].Weight != values[j].Weight {
			return values[i].Weight > values[j].Weight
		}
		return values[i].Code < values[j].Code
	})
	var sum uint64
	seen := make(map[string]struct{}, len(values))
	for index := range values {
		values[index].Code = strings.TrimSpace(values[index].Code)
		values[index].Title = strings.TrimSpace(values[index].Title)
		if values[index].Code == "" || values[index].Title == "" || values[index].Weight == 0 {
			return weightedCategory{}, fmt.Errorf("category code/title/weight are required")
		}
		if _, ok := seen[values[index].Code]; ok {
			return weightedCategory{}, fmt.Errorf("duplicate category %q", values[index].Code)
		}
		seen[values[index].Code] = struct{}{}
		sum += uint64(values[index].Weight)
	}
	if sum != 100 {
		return weightedCategory{}, fmt.Errorf("category weights sum to %d, want 100", sum)
	}
	return values[0], nil
}

func selectTopMonth(values []weightedMonth) (weightedMonth, error) {
	if len(values) == 0 {
		return weightedMonth{}, fmt.Errorf("months are required")
	}
	values = append([]weightedMonth(nil), values...)
	sort.Slice(values, func(i, j int) bool {
		if values[i].Weight != values[j].Weight {
			return values[i].Weight > values[j].Weight
		}
		return values[i].Month < values[j].Month
	})
	var sum uint64
	seen := make(map[uint32]struct{}, len(values))
	for _, value := range values {
		if value.Month < 1 || value.Month > 12 || value.Weight == 0 {
			return weightedMonth{}, fmt.Errorf("invalid month %d/weight %d", value.Month, value.Weight)
		}
		if _, ok := seen[value.Month]; ok {
			return weightedMonth{}, fmt.Errorf("duplicate month %d", value.Month)
		}
		seen[value.Month] = struct{}{}
		sum += uint64(value.Weight)
	}
	if sum != 100 {
		return weightedMonth{}, fmt.Errorf("month weights sum to %d, want 100", sum)
	}
	return values[0], nil
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

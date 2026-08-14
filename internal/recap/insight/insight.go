package insight

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/year-recap/internal/recap/model"
)

const (
	MaxTitleRunes       = 80
	MaxDescriptionRunes = 480
	MaxHighlightRunes   = 160
	MaxHighlights       = 5
)

type MetricFacts struct {
	TotalEvents        uint64  `json:"totalEvents"`
	Searches           uint64  `json:"searches"`
	TotalViews         uint64  `json:"totalViews"`
	UniqueListings     uint64  `json:"uniqueListings"`
	RepeatedViews      uint64  `json:"repeatedViews"`
	FavoritesAdded     uint64  `json:"favoritesAdded"`
	ChatsStarted       uint64  `json:"chatsStarted"`
	ListingsPublished  uint64  `json:"listingsPublished"`
	PurchasesCompleted uint64  `json:"purchasesCompleted"`
	SalesCompleted     uint64  `json:"salesCompleted"`
	ActiveDays         uint64  `json:"activeDays"`
	CategoriesCount    uint64  `json:"categoriesCount"`
	TopCategory        string  `json:"topCategory,omitempty"`
	TopCategoryViews   uint64  `json:"topCategoryViews"`
	RepeatRatePercent  float64 `json:"repeatRatePercent"`
	PurchaseRatePct    float64 `json:"purchaseRatePercent"`
}

type Facts struct {
	ProfileCode string      `json:"profileCode"`
	StartAt     time.Time   `json:"startAt"`
	EndAt       time.Time   `json:"endAt"`
	Metrics     MetricFacts `json:"metrics"`
}

type Card struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Highlights  []string `json:"highlights,omitempty"`
}

type Result struct {
	ProfileCode string    `json:"profileCode"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	Card        Card      `json:"card"`
	Facts       Facts     `json:"facts"`
}

type Generator interface {
	Generate(ctx context.Context, facts Facts) (Card, error)
}

func BuildFacts(profileCode string, start, end time.Time, metrics model.Metrics) Facts {
	return Facts{
		ProfileCode: profileCode,
		StartAt:     start.UTC(),
		EndAt:       end.UTC(),
		Metrics: MetricFacts{
			TotalEvents: metrics.TotalEvents, Searches: metrics.Searches,
			TotalViews: metrics.TotalViews, UniqueListings: metrics.UniqueListings,
			RepeatedViews: metrics.RepeatedViews, FavoritesAdded: metrics.FavoritesAdded,
			ChatsStarted: metrics.ChatsStarted, ListingsPublished: metrics.ListingsPublished,
			PurchasesCompleted: metrics.PurchasesCompleted, SalesCompleted: metrics.SalesCompleted,
			ActiveDays: metrics.ActiveDays, CategoriesCount: metrics.CategoriesCount,
			TopCategory: metrics.TopCategory, TopCategoryViews: metrics.TopCategoryViews,
			RepeatRatePercent: metrics.RepeatRate * 100, PurchaseRatePct: metrics.PurchaseRate * 100,
		},
	}
}

func ValidateCard(card Card) error {
	card.Title = strings.TrimSpace(card.Title)
	card.Description = strings.TrimSpace(card.Description)
	if card.Title == "" {
		return fmt.Errorf("behavior insight title is required")
	}
	if utf8.RuneCountInString(card.Title) > MaxTitleRunes {
		return fmt.Errorf("behavior insight title exceeds %d runes", MaxTitleRunes)
	}
	if card.Description == "" {
		return fmt.Errorf("behavior insight description is required")
	}
	if utf8.RuneCountInString(card.Description) > MaxDescriptionRunes {
		return fmt.Errorf("behavior insight description exceeds %d runes", MaxDescriptionRunes)
	}
	if len(card.Highlights) > MaxHighlights {
		return fmt.Errorf("behavior insight returned %d highlights, expected at most %d", len(card.Highlights), MaxHighlights)
	}
	for index, highlight := range card.Highlights {
		trimmed := strings.TrimSpace(highlight)
		if trimmed == "" {
			return fmt.Errorf("behavior insight highlight %d is empty", index)
		}
		if utf8.RuneCountInString(trimmed) > MaxHighlightRunes {
			return fmt.Errorf("behavior insight highlight %d exceeds %d runes", index, MaxHighlightRunes)
		}
	}
	return nil
}

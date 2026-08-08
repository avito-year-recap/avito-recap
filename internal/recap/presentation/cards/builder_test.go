package cards

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/share"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestBuildCreatesOrderedStory(t *testing.T) {
	configured := ruleset.DefaultRuleset()
	metrics := analytics.EnrichMetrics(model.Metrics{TotalEvents: 243, Searches: 20, TotalViews: 180, UniqueListings: 130, RepeatedViews: 50, FavoritesAdded: 30, ChatsStarted: 3, ChatsWithPurchase: 1, PurchasesCompleted: 1, ActiveDays: 45, CategoriesCount: 4, TopCategoryCode: "electronics", TopCategory: "Электроника", TopCategoryViews: 80, TopCategoryShareable: true, MostActiveMonth: 10})
	detected := behavior.Detect(configured, metrics)
	achievements := achievement.Build(configured, metrics)
	action := nextaction.Build(configured, metrics, model.ActionableState{FavoritesCount: 5, HasEverPublishedListing: true}, detected)
	shareID := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	public := share.Build(configured.SharePolicy, shareID, 2025, metrics, detected, achievements)
	values := Build(model.Profile{ID: uuid.MustParse("11111111-1111-4111-8111-111111111111"), Code: "buyer", DisplayName: "Алексей", Description: "Тест"}, 2025, metrics, detected, achievements, action, public)
	if len(values) < 7 {
		t.Fatalf("too few cards: %+v", values)
	}
	for i, card := range values {
		if card.Position != uint32(i+1) {
			t.Fatalf("position %d: %+v", i, card)
		}
	}
	if values[0].Type != model.CardIntro || values[len(values)-1].Type != model.CardShare {
		t.Fatalf("invalid boundaries: %+v", values)
	}
}

func TestMonthName(t *testing.T) {
	if MonthName(1) != "январь" || MonthName(12) != "декабрь" || MonthName(13) != "" {
		t.Fatal("month mapping")
	}
}

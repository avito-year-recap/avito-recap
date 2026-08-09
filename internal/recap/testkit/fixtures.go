package testkit

import (
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/achievement"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/behavior"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/nextaction"
	"github.com/year-recap/internal/recap/presentation/cards"
	"github.com/year-recap/internal/recap/ruleset"
)

var (
	ProfileID      = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	RecapID        = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	ShareID        = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	DraftListingID = uuid.MustParse("33333333-3333-4333-8333-333333333333")
)

func Clock() time.Time { return time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC) }

func Profile() model.Profile {
	return model.Profile{ID: ProfileID, Code: "active-buyer", DisplayName: "Алексей", Description: "Тестовый профиль"}
}

func Metrics() model.Metrics {
	return model.Metrics{
		TotalEvents: 243, Searches: 20, TotalViews: 180, UniqueListings: 130, RepeatedViews: 50,
		FavoritesAdded: 30, ChatsStarted: 3, ChatsWithPurchase: 1, PurchasesCompleted: 1,
		ActiveDays: 45, CategoriesCount: 4, TopCategoryCode: "electronics", TopCategory: "Электроника",
		TopCategoryViews: 80, TopCategoryShareable: true, MostActiveMonth: 10,
	}
}

func ActionableState() model.ActionableState {
	return model.ActionableState{CapturedAt: Clock(), FavoritesCount: 5, HasEverPublishedListing: true}
}

func Period() model.RecapPeriod {
	value, err := analytics.CompletedYearPeriod(2025, Clock())
	if err != nil {
		panic(err)
	}
	return value
}

func Recap() model.Recap {
	configured := ruleset.DefaultRuleset()
	metrics := analytics.EnrichMetrics(Metrics())
	state := ActionableState()
	detected := behavior.DetectWithRuleset(configured, metrics)
	achievements := achievement.BuildWithRuleset(configured, metrics)
	action := nextaction.BuildWithRuleset(configured, metrics, state, detected)
	return model.Recap{
		ID: RecapID, ShareID: ShareID, Profile: Profile(), Year: 2025, Period: Period(),
		RulesVersion: configured.Version, RulesDigest: configured.Digest(), Metrics: metrics,
		ActionableState: state, Behavior: detected, Achievements: achievements,
		Cards:      cards.BuildWithRuleset(configured, Profile(), 2025, ShareID, metrics, detected, achievements, action),
		NextAction: action, GeneratedAt: Clock(),
	}
}

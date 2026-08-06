package recap

import (
	"slices"
	"testing"
)

func TestBuildAchievementsForMVPProfiles(t *testing.T) {
	tests := []struct {
		name       string
		metrics    Metrics
		expected   []AchievementCode
		categories []AchievementCategory
	}{
		{
			name: "active buyer", metrics: Metrics{TotalViews: 180, FavoritesAdded: 28},
			expected:   []AchievementCode{AchievementAttentiveResearcher, AchievementMasterOfFavorites},
			categories: []AchievementCategory{AchievementCategoryDiscovery, AchievementCategoryCollection},
		},
		{
			name: "active seller", metrics: Metrics{ListingsPublished: 9, SalesCompleted: 6},
			expected:   []AchievementCode{AchievementSuccessfulSeller},
			categories: []AchievementCategory{AchievementCategorySelling},
		},
		{
			name: "researcher", metrics: Metrics{TotalViews: 260, CategoriesCount: 7},
			expected:   []AchievementCode{AchievementBroadInterests},
			categories: []AchievementCategory{AchievementCategoryDiscovery},
		},
		{
			name: "universal", metrics: Metrics{PurchasesCompleted: 1, SalesCompleted: 2, ListingsPublished: 4, ChatsStarted: 9},
			expected:   []AchievementCode{AchievementAllRounder},
			categories: []AchievementCategory{AchievementCategoryVersatility},
		},
		{
			name: "draft seller", metrics: Metrics{ListingsCreated: 7, ListingsPublished: 2},
			expected:   []AchievementCode{AchievementFirstSellingSteps},
			categories: []AchievementCategory{AchievementCategorySelling},
		},
		{
			name: "decisive buyer", metrics: Metrics{ChatsStarted: 15, ChatsWithPurchase: 4, PurchasesCompleted: 4},
			expected:   []AchievementCode{AchievementQuickDecision},
			categories: []AchievementCategory{AchievementCategoryBuying},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := BuildAchievements(test.metrics)
			assertAchievementCodes(t, actual, test.expected)
			for index, achievement := range actual {
				if achievement.Category != test.categories[index] {
					t.Fatalf("achievement %d category = %s, want %s", index, achievement.Category, test.categories[index])
				}
				if achievement.Priority == 0 || achievement.Title == "" || achievement.Reason == "" {
					t.Fatalf("incomplete achievement: %+v", achievement)
				}
			}
		})
	}
}

func TestBuildAchievementsBuyingGradesRemainReachable(t *testing.T) {
	t.Run("quick decision outranks broad deal closer", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			PurchasesCompleted: 4,
			ChatsStarted:       15,
			ChatsWithPurchase:  4,
		})
		assertAchievementCodes(t, result, []AchievementCode{AchievementQuickDecision})
	})

	t.Run("deal closer remains available without conversion evidence", func(t *testing.T) {
		result := BuildAchievements(Metrics{
			PurchasesCompleted: 4,
			ChatsStarted:       20,
			ChatsWithPurchase:  2,
		})
		assertAchievementCodes(t, result, []AchievementCode{AchievementDealCloser})
	})
}

func TestBuildAchievementsAwardsAtMostThreeDifferentCategories(t *testing.T) {
	result := BuildAchievements(allAchievementMetrics())

	expected := []AchievementCode{
		AchievementSuccessfulSeller,
		AchievementQuickDecision,
		AchievementBroadInterests,
	}
	assertAchievementCodes(t, result, expected)
	if len(result) != maxAchievements {
		t.Fatalf("achievement count = %d, want %d", len(result), maxAchievements)
	}
	seenCategories := make(map[AchievementCategory]struct{}, len(result))
	for index, achievement := range result {
		if _, exists := seenCategories[achievement.Category]; exists {
			t.Fatalf("category %s awarded twice: %+v", achievement.Category, result)
		}
		seenCategories[achievement.Category] = struct{}{}
		if index > 0 && !achievementLess(result[index-1], achievement) {
			t.Fatalf("achievements are not in deterministic order: %+v", result)
		}
	}
}

func TestBuildAchievementsKeepsHighestPriorityGradeInsideCategory(t *testing.T) {
	result := BuildAchievements(Metrics{ListingsCreated: 10, ListingsPublished: 9, SalesCompleted: 7})
	assertAchievementCodes(t, result, []AchievementCode{AchievementSuccessfulSeller})
	if result[0].Category != AchievementCategorySelling {
		t.Fatalf("category = %s", result[0].Category)
	}
}

func TestBuildAchievementsUsesCodeAsExplicitTieBreak(t *testing.T) {
	ruleset := DefaultRuleset()
	for index := range ruleset.AchievementPolicy.Rules {
		rule := &ruleset.AchievementPolicy.Rules[index]
		switch rule.Code {
		case AchievementSuccessfulSeller, AchievementConsistentPublisher:
			rule.Priority = 110 // intra-category tie: CONSISTENT_* wins by code.
		case AchievementDealCloser:
			rule.Priority = 110 // global tie: DEAL_* sorts before SUCCESSFUL/CONSISTENT.
		}
	}
	if err := ruleset.Validate(); err != nil {
		t.Fatal(err)
	}

	result := ruleset.BuildAchievements(allAchievementMetrics())
	assertAchievementCodes(t, result, []AchievementCode{
		AchievementConsistentPublisher,
		AchievementDealCloser,
		AchievementBroadInterests,
	})
}

func TestBuildAchievementsDoesNotDependOnPolicySliceOrder(t *testing.T) {
	ruleset := DefaultRuleset()
	forward := ruleset.BuildAchievements(allAchievementMetrics())
	forwardDigest := ruleset.Digest()
	slices.Reverse(ruleset.AchievementPolicy.Rules)
	backward := ruleset.BuildAchievements(allAchievementMetrics())
	if !equalAchievements(forward, backward) {
		t.Fatalf("policy order changed awards:\nforward: %+v\nreverse: %+v", forward, backward)
	}
	if ruleset.Digest() != forwardDigest {
		t.Fatal("semantically irrelevant policy order changed rules digest")
	}
}

func TestFullRecapProjectionContainsOnlySelectedThreeAchievements(t *testing.T) {
	ruleset := DefaultRuleset()
	metrics := EnrichMetrics(allAchievementMetrics())
	state := validActionableState()
	behavior := ruleset.DetectBehavior(metrics)
	achievements := ruleset.BuildAchievements(metrics)
	action := ruleset.BuildNextAction(metrics, state, behavior)
	cards := BuildCardsWithRuleset(ruleset, validProfile(), 2025, testShareID, metrics, behavior, achievements, action)

	card := findCard(t, cards, CardAchievement)
	payload := card.Payload.(AchievementPayload)
	if len(payload.Codes) != maxAchievements {
		t.Fatalf("achievement card contains %d codes, want %d", len(payload.Codes), maxAchievements)
	}
	for index, code := range payload.Codes {
		if code != achievements[index].Code {
			t.Fatalf("card code %d = %s, want %s", index, code, achievements[index].Code)
		}
	}

	value := Recap{
		ID: testRecapID, ShareID: testShareID, Profile: validProfile(), Year: 2025,
		Period: validPeriod(), RulesVersion: ruleset.Version, RulesDigest: ruleset.Digest(),
		Metrics: metrics, ActionableState: state, Behavior: behavior,
		Achievements: achievements, Cards: cards, NextAction: action, GeneratedAt: fixedClock(),
	}
	if err := validateRecapAgainstRuleset(value, ruleset, fixedClock()); err != nil {
		t.Fatalf("full recap with selected awards is invalid: %v", err)
	}
}

func TestBuildAchievementsBelowThresholds(t *testing.T) {
	result := BuildAchievements(Metrics{
		TotalViews: 149, FavoritesAdded: 19, ListingsCreated: 2,
		ListingsPublished: 2, CategoriesCount: 5,
	})
	if len(result) != 0 {
		t.Fatalf("expected no achievements, got %+v", result)
	}
}

func allAchievementMetrics() Metrics {
	return Metrics{
		TotalEvents: 392, TotalViews: 300, UniqueListings: 250, RepeatedViews: 50,
		FavoritesAdded: 40, ChatsStarted: 20, ChatsWithPurchase: 5,
		PurchasesCompleted: 5, ListingsCreated: 10, ListingsPublished: 10,
		SalesCompleted: 7, ActiveDays: 100, CategoriesCount: 8, MostActiveMonth: 1,
	}
}

func assertAchievementCodes(t *testing.T, actual []Achievement, expected []AchievementCode) {
	t.Helper()
	if len(actual) != len(expected) {
		t.Fatalf("achievement count = %d, want %d: %+v", len(actual), len(expected), actual)
	}
	for index, code := range expected {
		if actual[index].Code != code {
			t.Fatalf("achievement %d = %s, want %s", index, actual[index].Code, code)
		}
	}
}

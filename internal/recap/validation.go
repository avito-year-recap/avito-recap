package recap

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile = errors.New("invalid profile")
	ErrInvalidMetrics = errors.New("invalid metrics")
	ErrInvalidRecap   = errors.New("invalid recap")
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
	metrics = normalizeMetrics(metrics)

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

	// Counters for different action types are not a closed annual funnel.
	if metrics.ChatsWithPurchase > metrics.ChatsStarted {
		return fmt.Errorf("%w: chats with purchase exceed started chats", ErrInvalidMetrics)
	}

	if metrics.TopCategoryViews > metrics.TotalViews {
		return fmt.Errorf("%w: top-category views exceed total views", ErrInvalidMetrics)
	}

	categoryCode := metrics.TopCategoryCode
	categoryTitle := metrics.TopCategory
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

func validateRecap(value Recap) error {
	if value.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", ErrInvalidRecap)
	}
	if err := validateProfile(value.Profile); err != nil {
		return fmt.Errorf("%w: profile: %v", ErrInvalidRecap, err)
	}
	if value.Year == 0 {
		return fmt.Errorf("%w: year is required", ErrInvalidRecap)
	}
	if strings.TrimSpace(value.RulesVersion) == "" {
		return fmt.Errorf("%w: rules version is required", ErrInvalidRecap)
	}
	if err := validateMetrics(value.Metrics); err != nil {
		return fmt.Errorf("%w: metrics: %v", ErrInvalidRecap, err)
	}
	if value.Metrics.TotalEvents < minEventsForRecap {
		return fmt.Errorf("%w: total events are below recap minimum", ErrInvalidRecap)
	}
	if err := validateStoredRates(value.Metrics); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecap, err)
	}
	if !isKnownBehaviorCode(value.Behavior.Code) {
		return fmt.Errorf("%w: unknown behavior code %q", ErrInvalidRecap, value.Behavior.Code)
	}
	if value.Behavior.Title == "" || value.Behavior.Description == "" || value.Behavior.Reason == "" {
		return fmt.Errorf("%w: behavior text is incomplete", ErrInvalidRecap)
	}
	if len(value.Achievements) > maxAchievements {
		return fmt.Errorf("%w: too many achievements", ErrInvalidRecap)
	}
	seenAchievements := make(map[AchievementCode]struct{}, len(value.Achievements))
	for index, achievement := range value.Achievements {
		if !isKnownAchievementCode(achievement.Code) {
			return fmt.Errorf("%w: achievement %d has unknown code %q", ErrInvalidRecap, index, achievement.Code)
		}
		if achievement.Title == "" || achievement.Description == "" || achievement.Reason == "" {
			return fmt.Errorf("%w: achievement %d text is incomplete", ErrInvalidRecap, index)
		}
		if _, exists := seenAchievements[achievement.Code]; exists {
			return fmt.Errorf("%w: duplicate achievement code %q", ErrInvalidRecap, achievement.Code)
		}
		seenAchievements[achievement.Code] = struct{}{}
	}
	if !isKnownActionCode(value.NextAction.Code) {
		return fmt.Errorf("%w: unknown next-action code %q", ErrInvalidRecap, value.NextAction.Code)
	}
	if value.NextAction.Title == "" || value.NextAction.Description == "" ||
		value.NextAction.ButtonText == "" || value.NextAction.Reason == "" {
		return fmt.Errorf("%w: next-action text is incomplete", ErrInvalidRecap)
	}
	if len(value.Cards) == 0 {
		return fmt.Errorf("%w: cards are required", ErrInvalidRecap)
	}
	seenCardIDs := make(map[string]struct{}, len(value.Cards))
	for index, card := range value.Cards {
		if !isKnownCardType(card.Type) {
			return fmt.Errorf("%w: card %d has unknown type %q", ErrInvalidRecap, index, card.Type)
		}
		if card.ID == "" || card.Title == "" || card.Description == "" {
			return fmt.Errorf("%w: card %d is incomplete", ErrInvalidRecap, index)
		}
		if card.Position != uint32(index+1) {
			return fmt.Errorf("%w: card %q has position %d, want %d", ErrInvalidRecap, card.ID, card.Position, index+1)
		}
		if _, exists := seenCardIDs[card.ID]; exists {
			return fmt.Errorf("%w: duplicate card id %q", ErrInvalidRecap, card.ID)
		}
		seenCardIDs[card.ID] = struct{}{}
		if card.Payload.BehaviorCode != "" && !isKnownBehaviorCode(card.Payload.BehaviorCode) {
			return fmt.Errorf("%w: card %q has unknown behavior code", ErrInvalidRecap, card.ID)
		}
		if card.Payload.AchievementCode != "" && !isKnownAchievementCode(card.Payload.AchievementCode) {
			return fmt.Errorf("%w: card %q has unknown achievement code", ErrInvalidRecap, card.ID)
		}
		for _, code := range card.Payload.AchievementCodes {
			if !isKnownAchievementCode(code) {
				return fmt.Errorf("%w: card %q has unknown achievement code", ErrInvalidRecap, card.ID)
			}
		}
		if card.Payload.ActionCode != "" && !isKnownActionCode(card.Payload.ActionCode) {
			return fmt.Errorf("%w: card %q has unknown action code", ErrInvalidRecap, card.ID)
		}
	}
	if value.GeneratedAt.IsZero() {
		return fmt.Errorf("%w: generated time is required", ErrInvalidRecap)
	}

	return nil
}

func validateStoredRates(metrics Metrics) error {
	expected := EnrichMetrics(metrics)
	checks := []struct {
		name     string
		actual   float64
		expected float64
	}{
		{name: "repeat rate", actual: metrics.RepeatRate, expected: expected.RepeatRate},
		{name: "purchase rate", actual: metrics.PurchaseRate, expected: expected.PurchaseRate},
	}

	for _, check := range checks {
		if math.IsNaN(check.actual) || math.IsInf(check.actual, 0) ||
			math.Abs(check.actual-check.expected) > 1e-12 {
			return fmt.Errorf("%s is inconsistent: got %v, want %v", check.name, check.actual, check.expected)
		}
	}
	return nil
}

func isKnownBehaviorCode(code BehaviorCode) bool {
	switch code {
	case BehaviorActiveSeller, BehaviorStartingSeller, BehaviorDecisiveBuyer,
		BehaviorFindHunter, BehaviorResearcher, BehaviorUniversal:
		return true
	default:
		return false
	}
}

func isKnownAchievementCode(code AchievementCode) bool {
	switch code {
	case AchievementSuccessfulSeller, AchievementConsistentPublisher,
		AchievementAttentiveResearcher, AchievementMasterOfFavorites,
		AchievementBroadInterests, AchievementAllRounder,
		AchievementFirstSellingSteps, AchievementDealCloser,
		AchievementQuickDecision:
		return true
	default:
		return false
	}
}

func isKnownActionCode(code ActionCode) bool {
	switch code {
	case ActionFinishDraft, ActionOpenFavorites, ActionImproveListings,
		ActionContinueDialogs, ActionOpenTopCategory, ActionCreateFirstListing,
		ActionCreateListing, ActionSaveSearch, ActionViewSimilarListings,
		ActionExploreRecommendations:
		return true
	default:
		return false
	}
}

func isKnownCardType(cardType CardType) bool {
	switch cardType {
	case CardIntro, CardYearActivity, CardTopCategory, CardActiveMonth,
		CardBehavior, CardAchievement, CardMissedOpportunity,
		CardNextAction, CardSummary:
		return true
	default:
		return false
	}
}

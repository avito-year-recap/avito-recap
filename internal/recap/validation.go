package recap

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfile         = errors.New("invalid profile")
	ErrInvalidMetrics         = errors.New("invalid metrics")
	ErrInvalidActionableState = errors.New("invalid actionable state")
	ErrInvalidRecap           = errors.New("invalid recap")
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
	knownEvents, err := sumUint64(metrics.Searches, metrics.TotalViews, metrics.FavoritesAdded,
		metrics.ChatsStarted, metrics.ListingsCreated, metrics.ListingsPublished,
		metrics.PurchasesCompleted, metrics.SalesCompleted)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMetrics, err)
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
	if metrics.ActiveDays > metrics.TotalEvents && metrics.TotalEvents > 0 {
		return fmt.Errorf("%w: active days exceed total events", ErrInvalidMetrics)
	}
	if metrics.MostActiveMonth > 12 {
		return fmt.Errorf("%w: active month must be in range 0..12", ErrInvalidMetrics)
	}
	return nil
}

func validateMetricsForPeriod(metrics Metrics, period RecapPeriod) error {
	if err := validateMetrics(metrics); err != nil {
		return err
	}
	maxDays := uint64(period.EndAt.Sub(period.StartAt) / (24 * time.Hour))
	if maxDays == 0 || metrics.ActiveDays > maxDays {
		return fmt.Errorf("%w: active days %d exceed period length %d", ErrInvalidMetrics, metrics.ActiveDays, maxDays)
	}
	return nil
}

func validateActionableState(state ActionableState) error {
	if state.CapturedAt.IsZero() {
		return fmt.Errorf("%w: captured time is required", ErrInvalidActionableState)
	}
	if state.DraftListingID != uuid.Nil && state.CurrentDrafts == 0 {
		return fmt.Errorf("%w: draft id requires a positive draft count", ErrInvalidActionableState)
	}
	if state.OpenDialogID != uuid.Nil && state.OpenDialogs == 0 {
		return fmt.Errorf("%w: dialog id requires a positive open-dialog count", ErrInvalidActionableState)
	}
	if state.ActiveListingID != uuid.Nil && state.ActiveListings == 0 {
		return fmt.Errorf("%w: active listing id requires a positive active-listing count", ErrInvalidActionableState)
	}
	return nil
}

func validateRecap(value Recap) error {
	if value.ID == uuid.Nil {
		return fmt.Errorf("%w: internal id is required", ErrInvalidRecap)
	}
	if value.ShareID == uuid.Nil {
		return fmt.Errorf("%w: public share id is required", ErrInvalidRecap)
	}
	if value.ID == value.ShareID {
		return fmt.Errorf("%w: internal and public ids must differ", ErrInvalidRecap)
	}
	if err := validateProfile(value.Profile); err != nil {
		return fmt.Errorf("%w: profile: %v", ErrInvalidRecap, err)
	}
	if value.Year == 0 || value.Period.Year != value.Year {
		return fmt.Errorf("%w: year and period are inconsistent", ErrInvalidRecap)
	}
	if err := validatePeriod(value.Period); err != nil {
		return fmt.Errorf("%w: period: %v", ErrInvalidRecap, err)
	}
	if strings.TrimSpace(value.RulesVersion) == "" {
		return fmt.Errorf("%w: rules version is required", ErrInvalidRecap)
	}
	if err := validateMetricsForPeriod(value.Metrics, value.Period); err != nil {
		return fmt.Errorf("%w: metrics: %v", ErrInvalidRecap, err)
	}
	if value.Metrics.TotalEvents < minEventsForRecap {
		return fmt.Errorf("%w: total events are below recap minimum", ErrInvalidRecap)
	}
	if err := validateStoredRates(value.Metrics); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecap, err)
	}
	if err := validateActionableState(value.ActionableState); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecap, err)
	}
	if err := validateBehavior(value.Behavior); err != nil {
		return fmt.Errorf("%w: behavior: %v", ErrInvalidRecap, err)
	}
	if err := validateAchievements(value.Achievements); err != nil {
		return fmt.Errorf("%w: achievements: %v", ErrInvalidRecap, err)
	}
	if err := validateNextAction(value.NextAction); err != nil {
		return fmt.Errorf("%w: next action: %v", ErrInvalidRecap, err)
	}
	if err := validateCards(value.Cards); err != nil {
		return fmt.Errorf("%w: cards: %v", ErrInvalidRecap, err)
	}
	if err := validateShareCardConsistency(value); err != nil {
		return fmt.Errorf("%w: share card: %v", ErrInvalidRecap, err)
	}
	if value.GeneratedAt.IsZero() {
		return fmt.Errorf("%w: generated time is required", ErrInvalidRecap)
	}
	if value.GeneratedAt.Before(value.Period.EndAt) {
		return fmt.Errorf("%w: final recap was generated before period completion", ErrInvalidRecap)
	}
	if !value.ActionableState.CapturedAt.Equal(value.GeneratedAt) {
		return fmt.Errorf("%w: actionable-state capture time must equal generated time", ErrInvalidRecap)
	}
	return nil
}

func validatePeriod(period RecapPeriod) error {
	if period.Year == 0 || period.StartAt.IsZero() || period.EndAt.IsZero() || !period.Final {
		return errors.New("completed annual period is required")
	}
	expectedStart := time.Date(int(period.Year), time.January, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(int(period.Year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !period.StartAt.Equal(expectedStart) || !period.EndAt.Equal(expectedEnd) {
		return errors.New("period must cover exactly one UTC calendar year")
	}
	return nil
}

func validateBehavior(value Behavior) error {
	if !isKnownBehaviorCode(value.Code) {
		return fmt.Errorf("unknown code %q", value.Code)
	}
	if value.Title == "" || value.Description == "" || value.Reason == "" {
		return errors.New("text is incomplete")
	}
	if value.Code == BehaviorUniversal {
		if value.Score != 0 {
			return errors.New("universal behavior must have score 0")
		}
		return nil
	}
	if value.Score == 0 || len(value.Evidence) == 0 {
		return errors.New("scored behavior requires evidence")
	}
	var score uint32
	for index, item := range value.Evidence {
		if item.Metric == "" || item.Detail == "" || math.IsNaN(item.Actual) || math.IsInf(item.Actual, 0) ||
			math.IsNaN(item.Threshold) || math.IsInf(item.Threshold, 0) {
			return fmt.Errorf("evidence %d is invalid", index)
		}
		score += item.Points
	}
	if score != value.Score {
		return fmt.Errorf("evidence score %d differs from behavior score %d", score, value.Score)
	}
	return nil
}

func validateAchievements(values []Achievement) error {
	if len(values) > maxAchievements {
		return errors.New("too many achievements")
	}
	seen := make(map[AchievementCode]struct{}, len(values))
	for index, value := range values {
		if !isKnownAchievementCode(value.Code) {
			return fmt.Errorf("achievement %d has unknown code %q", index, value.Code)
		}
		if value.Title == "" || value.Description == "" || value.Reason == "" {
			return fmt.Errorf("achievement %d text is incomplete", index)
		}
		if _, ok := seen[value.Code]; ok {
			return fmt.Errorf("duplicate achievement code %q", value.Code)
		}
		seen[value.Code] = struct{}{}
	}
	return nil
}

func validateNextAction(value NextAction) error {
	if !isKnownActionCode(value.Code) {
		return fmt.Errorf("unknown code %q", value.Code)
	}
	if value.Title == "" || value.Description == "" || value.ButtonText == "" || value.Reason == "" {
		return errors.New("text is incomplete")
	}
	if err := validateActionTarget(value.Target); err != nil {
		return err
	}
	return validateTargetForAction(value.Code, value.Target)
}

func validateActionTarget(target ActionTarget) error {
	count := 0
	if target.Route != nil {
		count++
		if target.Route.Route == "" || target.Route.Route[0] != '/' {
			return errors.New("route target must contain an absolute application route")
		}
	}
	if target.Category != nil {
		count++
		if target.Category.CategoryCode == "" {
			return errors.New("category target code is required")
		}
	}
	if target.Listing != nil {
		count++
		if target.Listing.ListingID == uuid.Nil {
			return errors.New("listing target id is required")
		}
	}
	if target.Dialog != nil {
		count++
		if target.Dialog.DialogID == uuid.Nil {
			return errors.New("dialog target id is required")
		}
	}
	if target.Search != nil {
		count++
		if target.Search.CategoryCode == "" {
			return errors.New("search target category is required")
		}
	}
	if count != 1 {
		return fmt.Errorf("action target must contain exactly one destination, got %d", count)
	}
	return nil
}

func validateTargetForAction(code ActionCode, target ActionTarget) error {
	switch code {
	case ActionFinishDraft, ActionImproveListings, ActionViewSimilarListings:
		if target.Listing == nil {
			return fmt.Errorf("action %s requires a listing target", code)
		}
	case ActionContinueDialogs:
		if target.Dialog == nil {
			return fmt.Errorf("action %s requires a dialog target", code)
		}
	case ActionOpenTopCategory:
		if target.Category == nil {
			return fmt.Errorf("action %s requires a category target", code)
		}
	case ActionSaveSearch:
		if target.Search == nil {
			return fmt.Errorf("action %s requires a search target", code)
		}
	case ActionOpenFavorites, ActionCreateFirstListing, ActionCreateListing, ActionExploreRecommendations:
		if target.Route == nil {
			return fmt.Errorf("action %s requires a route target", code)
		}
	}
	return nil
}

func validateCards(cards []Card) error {
	if len(cards) == 0 {
		return errors.New("cards are required")
	}
	seen := make(map[string]struct{}, len(cards))
	shareCards := 0
	for index, card := range cards {
		if !isKnownCardType(card.Type) {
			return fmt.Errorf("card %d has unknown type %q", index, card.Type)
		}
		if card.ID == "" || card.Title == "" || card.Description == "" {
			return fmt.Errorf("card %d is incomplete", index)
		}
		if card.Position != uint32(index+1) {
			return fmt.Errorf("card %q has position %d, want %d", card.ID, card.Position, index+1)
		}
		if _, ok := seen[card.ID]; ok {
			return fmt.Errorf("duplicate card id %q", card.ID)
		}
		seen[card.ID] = struct{}{}
		if card.Type == CardShare {
			shareCards++
			if index != len(cards)-1 {
				return errors.New("share card must be the final story card")
			}
			if !card.Shareable {
				return errors.New("share card must be marked shareable")
			}
		} else if card.Shareable {
			return fmt.Errorf("only the final share card may be marked shareable, got %q", card.ID)
		}
		if err := validateCardPayload(card.Type, card.Payload); err != nil {
			return fmt.Errorf("card %q: %v", card.ID, err)
		}
	}
	if shareCards != 1 {
		return fmt.Errorf("exactly one share card is required, got %d", shareCards)
	}
	return nil
}

func validateCardPayload(cardType CardType, payload CardPayload) error {
	switch cardType {
	case CardIntro:
		if payload != nil {
			return errors.New("card must not have a payload")
		}
	case CardShare:
		value, ok := payload.(ShareCard)
		if !ok {
			return errors.New("requires share-card payload")
		}
		if err := validateShareCard(value); err != nil {
			return err
		}
	case CardYearActivity:
		if _, ok := payload.(YearActivityPayload); !ok {
			return errors.New("requires year-activity payload")
		}
	case CardTopCategory:
		value, ok := payload.(TopCategoryPayload)
		if !ok || value.CategoryCode == "" || value.Category == "" || value.CategoryViews == 0 {
			return errors.New("requires complete top-category payload")
		}
	case CardActiveMonth:
		value, ok := payload.(ActiveMonthPayload)
		if !ok || value.Month < 1 || value.Month > 12 {
			return errors.New("requires valid active-month payload")
		}
	case CardBehavior:
		value, ok := payload.(BehaviorPayload)
		if !ok || !isKnownBehaviorCode(value.Code) {
			return errors.New("requires valid behavior payload")
		}
	case CardAchievement:
		value, ok := payload.(AchievementPayload)
		if !ok || len(value.Codes) == 0 {
			return errors.New("requires achievement payload")
		}
		for _, code := range value.Codes {
			if !isKnownAchievementCode(code) {
				return errors.New("achievement payload has unknown code")
			}
		}
	case CardMissedOpportunity, CardNextAction:
		value, ok := payload.(ActionPayload)
		if !ok || !isKnownActionCode(value.Code) {
			return errors.New("requires action payload")
		}
		if err := validateActionTarget(value.Target); err != nil {
			return err
		}
		if err := validateTargetForAction(value.Code, value.Target); err != nil {
			return err
		}
	}
	return nil
}

func validateShareCard(value ShareCard) error {
	if value.ShareID == uuid.Nil {
		return errors.New("share id is required")
	}
	if value.Year == 0 {
		return errors.New("share year is required")
	}
	if strings.TrimSpace(value.BehaviorTitle) == "" {
		return errors.New("behavior title is required")
	}
	return nil
}

func validateShareCardConsistency(value Recap) error {
	if len(value.Cards) == 0 {
		return errors.New("cards are required")
	}
	last := value.Cards[len(value.Cards)-1]
	payload, ok := last.Payload.(ShareCard)
	if last.Type != CardShare || !ok {
		return errors.New("final card must contain a share-card payload")
	}
	expected := BuildShareCard(value)
	if payload != expected {
		return fmt.Errorf("stored payload %+v differs from public payload %+v", payload, expected)
	}
	return nil
}

func validateStoredRates(metrics Metrics) error {
	expected := EnrichMetrics(metrics)
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

func isKnownBehaviorCode(code BehaviorCode) bool {
	switch code {
	case BehaviorActiveSeller, BehaviorStartingSeller, BehaviorDecisiveBuyer, BehaviorFindHunter, BehaviorResearcher, BehaviorUniversal:
		return true
	default:
		return false
	}
}
func isKnownAchievementCode(code AchievementCode) bool {
	switch code {
	case AchievementSuccessfulSeller, AchievementConsistentPublisher, AchievementAttentiveResearcher, AchievementMasterOfFavorites, AchievementBroadInterests, AchievementAllRounder, AchievementFirstSellingSteps, AchievementDealCloser, AchievementQuickDecision:
		return true
	default:
		return false
	}
}
func isKnownActionCode(code ActionCode) bool {
	switch code {
	case ActionFinishDraft, ActionOpenFavorites, ActionImproveListings, ActionContinueDialogs, ActionOpenTopCategory, ActionCreateFirstListing, ActionCreateListing, ActionSaveSearch, ActionViewSimilarListings, ActionExploreRecommendations:
		return true
	default:
		return false
	}
}
func isKnownCardType(cardType CardType) bool {
	switch cardType {
	case CardIntro, CardYearActivity, CardTopCategory, CardActiveMonth, CardBehavior, CardAchievement, CardMissedOpportunity, CardNextAction, CardShare:
		return true
	default:
		return false
	}
}

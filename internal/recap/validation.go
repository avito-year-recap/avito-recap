package recap

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var categoryCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

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
	if metrics.ActiveDays > metrics.TotalEvents {
		return fmt.Errorf("%w: active days exceed total events", ErrInvalidMetrics)
	}
	if metrics.TopCategoryCode != "" && !isSafeCategoryCode(metrics.TopCategoryCode) {
		return fmt.Errorf("%w: top category code is unsafe", ErrInvalidMetrics)
	}
	if metrics.MostActiveMonth > 12 {
		return fmt.Errorf("%w: active month must be in range 0..12", ErrInvalidMetrics)
	}
	seenCategoryActivities := make(map[string]struct{}, len(metrics.CategoryActivities))
	var categoryViews, categoryFavorites, categoryPurchases uint64
	for index, activity := range metrics.CategoryActivities {
		if activity.CategoryCode == "" || activity.Category == "" {
			return fmt.Errorf("%w: category activity %d requires code and title", ErrInvalidMetrics, index)
		}
		if !isSafeCategoryCode(activity.CategoryCode) {
			return fmt.Errorf("%w: category activity %d has unsafe code", ErrInvalidMetrics, index)
		}
		if _, exists := seenCategoryActivities[activity.CategoryCode]; exists {
			return fmt.Errorf("%w: duplicate category activity %q", ErrInvalidMetrics, activity.CategoryCode)
		}
		seenCategoryActivities[activity.CategoryCode] = struct{}{}
		if activity.Views == 0 && activity.FavoritesAdded == 0 && activity.PurchasesCompleted == 0 {
			return fmt.Errorf("%w: category activity %q has no evidence", ErrInvalidMetrics, activity.CategoryCode)
		}
		var err error
		categoryViews, err = sumUint64(categoryViews, activity.Views)
		if err != nil {
			return fmt.Errorf("%w: category views overflow", ErrInvalidMetrics)
		}
		categoryFavorites, err = sumUint64(categoryFavorites, activity.FavoritesAdded)
		if err != nil {
			return fmt.Errorf("%w: category favorites overflow", ErrInvalidMetrics)
		}
		categoryPurchases, err = sumUint64(categoryPurchases, activity.PurchasesCompleted)
		if err != nil {
			return fmt.Errorf("%w: category purchases overflow", ErrInvalidMetrics)
		}
	}
	if uint64(len(metrics.CategoryActivities)) > metrics.CategoriesCount {
		return fmt.Errorf("%w: detailed category count exceeds categories count", ErrInvalidMetrics)
	}
	if categoryViews > metrics.TotalViews || categoryFavorites > metrics.FavoritesAdded || categoryPurchases > metrics.PurchasesCompleted {
		return fmt.Errorf("%w: per-category counters exceed aggregate counters", ErrInvalidMetrics)
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
	if metrics.TotalEvents > 0 && (metrics.MostActiveMonth < 1 || metrics.MostActiveMonth > 12) {
		return fmt.Errorf("%w: active month is required for a non-empty annual period", ErrInvalidMetrics)
	}
	return nil
}

func validateActionableState(state ActionableState) error {
	if state.CapturedAt.IsZero() {
		return fmt.Errorf("%w: captured time is required", ErrInvalidActionableState)
	}
	if (state.DraftListingID == uuid.Nil) != (state.CurrentDrafts == 0) {
		return fmt.Errorf("%w: draft count and addressable draft id must be present together", ErrInvalidActionableState)
	}
	if (state.OpenDialogID == uuid.Nil) != (state.OpenDialogs == 0) {
		return fmt.Errorf("%w: open-dialog count and addressable dialog id must be present together", ErrInvalidActionableState)
	}
	if (state.ActiveListingID == uuid.Nil) != (state.ActiveListings == 0) {
		return fmt.Errorf("%w: active-listing count and addressable listing id must be present together", ErrInvalidActionableState)
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
	if !semanticVersionPattern.MatchString(strings.TrimSpace(value.RulesVersion)) {
		return fmt.Errorf("%w: semantic rules version is required", ErrInvalidRecap)
	}
	if !isSHA256Hex(value.RulesDigest) {
		return fmt.Errorf("%w: rules digest is required", ErrInvalidRecap)
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
	var score uint64
	for index, item := range value.Evidence {
		if item.Metric == "" || item.Detail == "" || math.IsNaN(item.Actual) || math.IsInf(item.Actual, 0) ||
			math.IsNaN(item.Threshold) || math.IsInf(item.Threshold, 0) {
			return fmt.Errorf("evidence %d is invalid", index)
		}
		score += uint64(item.Points)
	}
	if score != uint64(value.Score) {
		return fmt.Errorf("evidence score %d differs from behavior score %d", score, value.Score)
	}
	return nil
}

func validateAchievements(values []Achievement) error {
	if len(values) > maxAchievements {
		return fmt.Errorf("too many achievements: got %d, maximum is %d", len(values), maxAchievements)
	}
	seenCodes := make(map[AchievementCode]struct{}, len(values))
	for index, value := range values {
		if !isKnownAchievementCode(value.Code) {
			return fmt.Errorf("achievement %d has unknown code %q", index, value.Code)
		}
		if !isKnownAchievementCategory(value.Category) {
			return fmt.Errorf("achievement %d has unknown category %q", index, value.Category)
		}
		if value.Title == "" || value.Description == "" || value.Reason == "" {
			return fmt.Errorf("achievement %d text is incomplete", index)
		}
		if _, ok := seenCodes[value.Code]; ok {
			return fmt.Errorf("duplicate achievement code %q", value.Code)
		}
		seenCodes[value.Code] = struct{}{}
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
		if !isSafeApplicationRoute(target.Route.Route) {
			return errors.New("route target must contain a known safe application route")
		}
	}
	if target.Category != nil {
		count++
		if !isSafeCategoryCode(target.Category.CategoryCode) {
			return errors.New("safe category target code is required")
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
		if !isSafeCategoryCode(target.Search.CategoryCode) {
			return errors.New("safe search target category is required")
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
	case ActionOpenFavorites:
		if target.Route == nil || target.Route.Route != "/favorites" {
			return fmt.Errorf("action %s requires /favorites route", code)
		}
	case ActionCreateFirstListing, ActionCreateListing:
		if target.Route == nil || target.Route.Route != "/listings/new" {
			return fmt.Errorf("action %s requires /listings/new route", code)
		}
	case ActionExploreRecommendations:
		if target.Route == nil || target.Route.Route != "/recommendations" {
			return fmt.Errorf("action %s requires /recommendations route", code)
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
		if !ok || len(value.Codes) == 0 || len(value.Codes) > maxAchievements {
			return errors.New("requires one to three achievement codes")
		}
		seen := make(map[AchievementCode]struct{}, len(value.Codes))
		for _, code := range value.Codes {
			if !isKnownAchievementCode(code) {
				return errors.New("achievement payload has unknown code")
			}
			if _, exists := seen[code]; exists {
				return errors.New("achievement payload has duplicate code")
			}
			seen[code] = struct{}{}
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
	if strings.TrimSpace(value.PrivacyVersion) == "" {
		return errors.New("privacy version is required")
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
	if payload.ShareID != value.ShareID || payload.Year != value.Year {
		return errors.New("share payload is not bound to recap share id and year")
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

func isSafeCategoryCode(code string) bool {
	return categoryCodePattern.MatchString(strings.TrimSpace(code))
}

func isSafeApplicationRoute(route string) bool {
	switch route {
	case "/favorites", "/listings/new", "/recommendations":
		return true
	default:
		return false
	}
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

func isKnownBehaviorCode(code BehaviorCode) bool {
	switch code {
	case BehaviorActiveSeller, BehaviorStartingSeller, BehaviorDecisiveBuyer, BehaviorFindHunter, BehaviorResearcher, BehaviorUniversal:
		return true
	default:
		return false
	}
}
func isKnownAchievementCategory(category AchievementCategory) bool {
	switch category {
	case AchievementCategorySelling, AchievementCategoryBuying, AchievementCategoryDiscovery,
		AchievementCategoryCollection, AchievementCategoryVersatility, AchievementCategoryInterest:
		return true
	default:
		return false
	}
}
func isKnownAchievementCode(code AchievementCode) bool {
	switch code {
	case AchievementSuccessfulSeller, AchievementConsistentPublisher, AchievementAttentiveResearcher,
		AchievementMasterOfFavorites, AchievementBroadInterests, AchievementAllRounder,
		AchievementFirstSellingSteps, AchievementDealCloser, AchievementQuickDecision,
		AchievementStyleIcon, AchievementFashionableMan, AchievementTraveler,
		AchievementForTheSoul, AchievementBookworm, AchievementBeautyConnoisseur,
		AchievementInTheRhythmOfMusic, AchievementWorldOfPlay, AchievementMasterCraft,
		AchievementCaringOwner, AchievementLittleDiscoveries:
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

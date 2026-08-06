package recap

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildCardsFullAchievementScenario(t *testing.T) {
	profile := validProfile()
	metrics := EnrichMetrics(validMetrics())
	behavior := DetectBehavior(metrics)
	achievements := BuildAchievements(metrics)
	nextAction := BuildNextAction(metrics, validActionableState())

	cards := BuildCards(profile, 2025, testShareID, metrics, behavior, achievements, nextAction)

	if len(cards) != 8 {
		t.Fatalf("expected 8 story cards, got %d: %+v", len(cards), cards)
	}
	assertCardSequence(t, cards, []CardType{
		CardIntro,
		CardYearActivity,
		CardTopCategory,
		CardActiveMonth,
		CardBehavior,
		CardAchievement,
		CardNextAction,
		CardShare,
	})

	achievementCard := findCard(t, cards, CardAchievement)
	achievementPayload, ok := achievementCard.Payload.(AchievementPayload)
	if !ok || len(achievementPayload.Codes) != len(achievements) {
		t.Fatalf("achievement payload = %+v, want %d codes", achievementCard.Payload, len(achievements))
	}
	if achievementCard.Shareable {
		t.Fatal("personal achievement card must not be directly shareable")
	}

	categoryCard := findCard(t, cards, CardTopCategory)
	if categoryCard.Shareable {
		t.Fatal("personal category card must not be directly shareable")
	}
	categoryPayload, ok := categoryCard.Payload.(TopCategoryPayload)
	if !ok || categoryPayload.CategoryCode != metrics.TopCategoryCode ||
		categoryPayload.Category != metrics.TopCategory ||
		categoryPayload.CategoryViews != metrics.TopCategoryViews {
		t.Fatalf("unexpected category payload: %+v", categoryCard.Payload)
	}

	share := findCard(t, cards, CardShare)
	sharePayload, ok := share.Payload.(ShareCard)
	if !ok {
		t.Fatalf("share card has wrong payload: %T", share.Payload)
	}
	if !share.Shareable || sharePayload.ShareID != testShareID || sharePayload.Year != 2025 ||
		sharePayload.BehaviorTitle != behavior.Title || sharePayload.AchievementTitle == "" ||
		sharePayload.TopCategory != metrics.TopCategory {
		t.Fatalf("unexpected final share card: %+v", share)
	}
}

func TestBuildCardsKeepsAchievementsWithMissedOpportunity(t *testing.T) {
	metrics := Metrics{
		TotalEvents:       209,
		TotalViews:        150,
		UniqueListings:    145,
		ChatsStarted:      3,
		ListingsCreated:   7,
		ListingsPublished: 2,
		CategoriesCount:   6,
		TopCategoryCode:   "electronics",
		TopCategory:       "Электроника",
		TopCategoryViews:  40,
		MostActiveMonth:   12,
	}
	metrics = EnrichMetrics(metrics)
	cards := BuildCards(
		validProfile(),
		2025,
		testShareID,
		metrics,
		DetectBehavior(metrics),
		BuildAchievements(metrics),
		BuildNextAction(metrics, ActionableState{CurrentDrafts: 1, DraftListingID: testDraftListingID}),
	)

	achievementCard := findCard(t, cards, CardAchievement)
	findCard(t, cards, CardMissedOpportunity)
	payload, ok := achievementCard.Payload.(AchievementPayload)
	if !ok {
		t.Fatalf("researcher achievement card has wrong payload: %T", achievementCard.Payload)
	}
	wantCodes := []AchievementCode{AchievementBroadInterests, AchievementFirstSellingSteps}
	if !reflect.DeepEqual(payload.Codes, wantCodes) {
		t.Fatalf("researcher achievements were hidden or changed: got %v, want %v", payload.Codes, wantCodes)
	}
	if len(cards) != 9 {
		t.Fatalf("expected achievement and missed-opportunity cards to coexist, got %d cards", len(cards))
	}
}

func TestBuildCardsOmitsUnavailableOptionalCards(t *testing.T) {
	cards := BuildCards(
		validProfile(),
		2025,
		testShareID,
		Metrics{TotalEvents: 5},
		DetectBehavior(Metrics{}),
		nil,
		BuildNextAction(Metrics{}, ActionableState{HasEverPublishedListing: true}),
	)

	for _, card := range cards {
		if card.Type == CardTopCategory || card.Type == CardActiveMonth ||
			card.Type == CardAchievement || card.Type == CardMissedOpportunity {
			t.Fatalf("optional card %s must be omitted", card.Type)
		}
	}
	if len(cards) != 5 {
		t.Fatalf("expected 5 base cards, got %d", len(cards))
	}
	if cards[len(cards)-1].Type != CardShare || !cards[len(cards)-1].Shareable {
		t.Fatalf("final card must be the share card: %+v", cards[len(cards)-1])
	}
}

func TestBuildCardsDoesNotShareSensitiveCategory(t *testing.T) {
	metrics := Metrics{
		TotalEvents:      5,
		TotalViews:       5,
		TopCategoryCode:  "sensitive",
		TopCategory:      "Чувствительная категория",
		TopCategoryViews: 5,
	}

	cards := BuildCards(
		validProfile(),
		2025,
		testShareID,
		metrics,
		DetectBehavior(metrics),
		nil,
		BuildNextAction(metrics, ActionableState{}),
	)

	sharePayload, ok := findCard(t, cards, CardShare).Payload.(ShareCard)
	if !ok {
		t.Fatal("share card payload is missing")
	}
	if sharePayload.TopCategory != "" {
		t.Fatalf("sensitive category leaked into share card: %+v", sharePayload)
	}
}

func TestMonthName(t *testing.T) {
	tests := map[uint32]string{0: "", 1: "январь", 6: "июнь", 12: "декабрь", 13: ""}
	for month, expected := range tests {
		if actual := monthName(month); actual != expected {
			t.Fatalf("monthName(%d) = %q, want %q", month, actual, expected)
		}
	}
}

func assertCardSequence(t *testing.T, cards []Card, expected []CardType) {
	t.Helper()
	if len(cards) != len(expected) {
		t.Fatalf("card count = %d, want %d", len(cards), len(expected))
	}
	seenIDs := make(map[string]struct{}, len(cards))
	for index, card := range cards {
		if card.Type != expected[index] {
			t.Fatalf("card %d type = %s, want %s", index, card.Type, expected[index])
		}
		if card.Position != uint32(index+1) {
			t.Fatalf("card %q position = %d, want %d", card.ID, card.Position, index+1)
		}
		if card.ID == "" || card.Title == "" || card.Description == "" {
			t.Fatalf("incomplete card at position %d: %+v", index, card)
		}
		if _, exists := seenIDs[card.ID]; exists {
			t.Fatalf("duplicate card id: %s", card.ID)
		}
		seenIDs[card.ID] = struct{}{}
	}
}

func findCard(t *testing.T, cards []Card, cardType CardType) Card {
	t.Helper()
	for _, card := range cards {
		if card.Type == cardType {
			return card
		}
	}
	t.Fatalf("card %s not found", cardType)
	return Card{}
}

func TestBuildCardsUsesCurrentDraftStateInsteadOfAnnualDifference(t *testing.T) {
	draftID := testDraftListingID
	metrics := Metrics{TotalEvents: 10, ListingsCreated: 3, ListingsPublished: 2}
	metrics = EnrichMetrics(metrics)
	action := BuildNextAction(metrics, ActionableState{CurrentDrafts: 1, DraftListingID: draftID})
	cards := BuildCards(validProfile(), 2025, testShareID, metrics, DetectBehavior(metrics), BuildAchievements(metrics), action)
	card := findCard(t, cards, CardMissedOpportunity)
	if strings.Contains(strings.ToLower(card.Explanation), "черновиков осталось") {
		t.Fatalf("card must not derive exact draft count from annual counters: %+v", card)
	}
	if !strings.Contains(card.Explanation, "актуальное состояние") {
		t.Fatalf("card must name current-state source: %q", card.Explanation)
	}
	payload, ok := card.Payload.(ActionPayload)
	if !ok || payload.Target.Listing == nil || payload.Target.Listing.ListingID != draftID {
		t.Fatalf("missed-opportunity target is not typed: %+v", card.Payload)
	}
}

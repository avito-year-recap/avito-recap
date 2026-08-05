package recap

import (
	"strings"
	"testing"
)

func TestBuildCardsFullAchievementScenario(t *testing.T) {
	profile := validProfile()
	metrics := EnrichMetrics(validMetrics())
	behavior := DetectBehavior(metrics)
	achievements := BuildAchievements(metrics)
	nextAction := BuildNextAction(metrics)

	cards := BuildCards(profile, 2025, metrics, behavior, achievements, nextAction)

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
		CardSummary,
	})

	achievementCard := findCard(t, cards, CardAchievement)
	if len(achievementCard.Payload.AchievementCodes) != len(achievements) {
		t.Fatalf("achievement payload = %+v, want %d codes", achievementCard.Payload, len(achievements))
	}
	if !achievementCard.Shareable {
		t.Fatal("card with only public achievements must be shareable")
	}

	categoryCard := findCard(t, cards, CardTopCategory)
	if !categoryCard.Shareable {
		t.Fatal("safe top category card must be shareable")
	}
	if categoryCard.Payload.CategoryCode != metrics.TopCategoryCode ||
		categoryCard.Payload.Category != metrics.TopCategory ||
		categoryCard.Payload.CategoryViews != metrics.TopCategoryViews {
		t.Fatalf("unexpected category payload: %+v", categoryCard.Payload)
	}
}

func TestBuildCardsKeepsAchievementsWithMissedOpportunity(t *testing.T) {
	metrics := Metrics{
		TotalEvents:      200,
		TotalViews:       150,
		UniqueListings:   145,
		ChatsStarted:     3,
		CategoriesCount:  6,
		TopCategoryCode:  "electronics",
		TopCategory:      "Электроника",
		TopCategoryViews: 40,
		MostActiveMonth:  12,
	}
	metrics = EnrichMetrics(metrics)
	cards := BuildCards(
		validProfile(),
		2025,
		metrics,
		DetectBehavior(metrics),
		BuildAchievements(metrics),
		BuildNextAction(metrics),
	)

	findCard(t, cards, CardAchievement)
	findCard(t, cards, CardMissedOpportunity)
	if len(cards) != 9 {
		t.Fatalf("expected achievement and missed-opportunity cards to coexist, got %d cards", len(cards))
	}
}

func TestBuildCardsOmitsUnavailableOptionalCards(t *testing.T) {
	cards := BuildCards(
		validProfile(),
		2025,
		Metrics{TotalEvents: 5},
		DetectBehavior(Metrics{}),
		nil,
		BuildNextAction(Metrics{}),
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
		metrics,
		DetectBehavior(metrics),
		nil,
		BuildNextAction(metrics),
	)

	if findCard(t, cards, CardTopCategory).Shareable {
		t.Fatal("category without explicit safety flag must not be shareable")
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

func TestBuildCardsDoesNotInventDraftCount(t *testing.T) {
	metrics := Metrics{
		TotalEvents:       10,
		ListingsCreated:   3,
		ListingsPublished: 2,
	}
	metrics = EnrichMetrics(metrics)
	cards := BuildCards(
		validProfile(),
		2025,
		metrics,
		DetectBehavior(metrics),
		BuildAchievements(metrics),
		BuildNextAction(metrics),
	)

	card := findCard(t, cards, CardMissedOpportunity)
	if strings.Contains(strings.ToLower(card.Explanation), "черновиков осталось") {
		t.Fatalf("card must not present annual counter difference as exact draft count: %+v", card)
	}
	if !strings.Contains(card.Explanation, "не показывают точное число") {
		t.Fatalf("card must explain the limitation of annual counters: %q", card.Explanation)
	}
}

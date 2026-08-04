package recap

import "testing"

func TestBuildCardsFullScenario(t *testing.T) {
	profile := validProfile()
	metrics := validMetrics()
	metrics = EnrichMetrics(metrics)
	behavior := DetectBehavior(metrics)
	achievements := BuildAchievements(metrics)
	nextAction := BuildNextAction(metrics)

	cards := BuildCards(profile, 2025, metrics, behavior, achievements, nextAction)

	expectedCount := 7 + len(achievements)
	if len(cards) != expectedCount {
		t.Fatalf("expected %d cards, got %d", expectedCount, len(cards))
	}

	seenIDs := make(map[string]struct{}, len(cards))
	for index, card := range cards {
		wantPosition := uint32(index + 1)
		if card.Position != wantPosition {
			t.Fatalf("card %q position = %d, want %d", card.ID, card.Position, wantPosition)
		}
		if card.ID == "" || card.Title == "" || card.Description == "" {
			t.Fatalf("incomplete card at position %d: %+v", index, card)
		}
		if _, exists := seenIDs[card.ID]; exists {
			t.Fatalf("duplicate card id: %s", card.ID)
		}
		seenIDs[card.ID] = struct{}{}
	}

	if cards[0].Type != CardIntro || cards[1].Type != CardYearActivity {
		t.Fatalf("unexpected first cards: %s, %s", cards[0].Type, cards[1].Type)
	}
	if cards[len(cards)-2].Type != CardNextAction || cards[len(cards)-1].Type != CardSummary {
		t.Fatalf("unexpected final cards: %s, %s", cards[len(cards)-2].Type, cards[len(cards)-1].Type)
	}

	categoryCard := findCard(t, cards, CardTopCategory)
	if !categoryCard.Shareable {
		t.Fatal("safe top category card must be shareable")
	}
	if categoryCard.Payload.Category != metrics.TopCategory || categoryCard.Payload.CategoryViews != metrics.TopCategoryViews {
		t.Fatalf("unexpected category payload: %+v", categoryCard.Payload)
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
		if card.Type == CardTopCategory || card.Type == CardActiveMonth || card.Type == CardAchievement {
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
	tests := map[uint32]string{
		0:  "",
		1:  "январь",
		6:  "июнь",
		12: "декабрь",
		13: "",
	}
	for month, expected := range tests {
		if actual := monthName(month); actual != expected {
			t.Fatalf("monthName(%d) = %q, want %q", month, actual, expected)
		}
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

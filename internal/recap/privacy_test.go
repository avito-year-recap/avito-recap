package recap

import "testing"

func TestBuildShareCardUsesOnlyExplicitlyPublicData(t *testing.T) {
	value := Recap{
		ID:      testRecapID,
		ShareID: testShareID,
		Year:    2025,
		Profile: Profile{
			ID:          testProfileID,
			DisplayName: "Private name",
		},
		Metrics: Metrics{
			TopCategory:          "Электроника",
			TopCategoryShareable: true,
		},
		Behavior: Behavior{Title: "Исследователь"},
		Achievements: []Achievement{
			{Title: "Private achievement", Shareable: false},
			{Title: "Public achievement", Shareable: true},
		},
	}

	actual := BuildShareCard(value)
	if actual.ShareID != value.ShareID || actual.Year != value.Year || actual.BehaviorTitle != value.Behavior.Title {
		t.Fatalf("unexpected base share data: %+v", actual)
	}
	if actual.AchievementTitle != "Public achievement" {
		t.Fatalf("unexpected achievement: %q", actual.AchievementTitle)
	}
	if actual.TopCategory != "Электроника" {
		t.Fatalf("unexpected top category: %q", actual.TopCategory)
	}
}

func TestBuildShareCardRedactsCategoryWithoutSafetyFlag(t *testing.T) {
	actual := BuildShareCard(Recap{
		Metrics: Metrics{TopCategory: "Sensitive", TopCategoryShareable: false},
	})
	if actual.TopCategory != "" {
		t.Fatalf("expected category to be redacted, got %q", actual.TopCategory)
	}
}

func TestBuildShareCardHandlesNoPublicAchievement(t *testing.T) {
	actual := BuildShareCard(Recap{
		Achievements: []Achievement{{Title: "Private", Shareable: false}},
	})
	if actual.AchievementTitle != "" {
		t.Fatalf("expected empty achievement, got %q", actual.AchievementTitle)
	}
}

func TestFinalStoryCardMatchesPublicShareCard(t *testing.T) {
	value := validRecap()
	last := value.Cards[len(value.Cards)-1]
	if last.Type != CardShare || !last.Shareable {
		t.Fatalf("unexpected final card: %+v", last)
	}
	payload, ok := last.Payload.(ShareCard)
	if !ok {
		t.Fatalf("final card payload = %T, want ShareCard", last.Payload)
	}
	if expected := BuildShareCard(value); payload != expected {
		t.Fatalf("final story payload = %+v, public payload = %+v", payload, expected)
	}
}

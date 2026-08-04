package recap

import "testing"

func TestBuildShareCardUsesOnlyExplicitlyPublicData(t *testing.T) {
	value := Recap{
		ID:   "11111111-1111-4111-8111-111111111111",
		Year: 2025,
		Profile: Profile{
			ID:          "private-profile-id",
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
	if actual.RecapID != value.ID || actual.Year != value.Year || actual.BehaviorTitle != value.Behavior.Title {
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

package share

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

func TestBuildRedactsPrivateData(t *testing.T) {
	value := model.Recap{ShareID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), Year: 2025, Metrics: model.Metrics{TopCategoryCode: "electronics", TopCategory: "Электроника", TopCategoryShareable: true}, Behavior: model.Behavior{Title: "Исследователь"}, Achievements: []model.Achievement{{Code: model.AchievementBroadInterests, Title: "Открытие года", Shareable: true}}}
	got := Build(value)
	if got.BehaviorTitle != "Исследователь" || got.AchievementTitle != "Открытие года" || got.TopCategory != "Электроника" {
		t.Fatalf("got %+v", got)
	}
}

func TestBuildDoesNotExposeThematicAchievement(t *testing.T) {
	got := Build(model.Recap{Achievements: []model.Achievement{{Code: model.AchievementBookworm, Title: "Книжный червь", Shareable: true}}})
	if got.AchievementTitle != "" {
		t.Fatalf("leaked: %+v", got)
	}
}

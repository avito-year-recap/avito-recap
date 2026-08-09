package share

import (
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func TestBuildRedactsPrivateData(t *testing.T) {
	value := model.Recap{ShareID: uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), Year: 2025, Metrics: model.Metrics{TopCategoryCode: "electronics", TopCategory: "Электроника", TopCategoryShareable: true}, Behavior: model.Behavior{Title: "Исследователь"}, Achievements: []model.Achievement{{Code: model.AchievementBroadInterests, Title: "Открытие года", Shareable: true}}}
	got := Build(ruleset.DefaultRuleset().SharePolicy, value.ShareID, value.Year, value.Metrics, value.Behavior, value.Achievements)
	if got.BehaviorTitle != "Исследователь" || got.AchievementTitle != "Открытие года" || got.TopCategory != "Электроника" {
		t.Fatalf("got %+v", got)
	}
}

func TestBuildDoesNotExposeThematicAchievement(t *testing.T) {
	got := Build(ruleset.DefaultRuleset().SharePolicy, uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"), 2025, model.Metrics{}, model.Behavior{}, []model.Achievement{{Code: model.AchievementBookworm, Title: "Книжный червь", Shareable: true}})
	if got.AchievementTitle != "" {
		t.Fatalf("leaked: %+v", got)
	}
}

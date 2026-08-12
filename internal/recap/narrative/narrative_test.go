package narrative_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/narrative"
)

func TestApplyChangesOnlyAllowListedCardDescriptions(t *testing.T) {
	recap := sampleRecap()
	beforePayload := recap.Cards[1].Payload
	beforeExplanation := recap.Cards[1].Explanation
	beforeTitle := recap.Cards[1].Title
	shareBefore := recap.Cards[2].Description

	enriched, err := narrative.Apply(recap, narrative.Story{Cards: []narrative.CardNarrative{
		{ID: "intro", Description: "Год начался с любопытства и сложился в личную историю."},
		{ID: "year-activity", Description: "Активность года сложилась из разных сценариев и интересов."},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if enriched.Cards[0].Description == recap.Cards[0].Description || enriched.Cards[1].Description == recap.Cards[1].Description {
		t.Fatal("editable descriptions were not enriched")
	}
	if enriched.Cards[1].Title != beforeTitle || enriched.Cards[1].Explanation != beforeExplanation || !reflect.DeepEqual(enriched.Cards[1].Payload, beforePayload) {
		t.Fatalf("AI changed protected card fields: %+v", enriched.Cards[1])
	}
	if enriched.Cards[2].Description != shareBefore {
		t.Fatalf("SHARE description changed: %q -> %q", shareBefore, enriched.Cards[2].Description)
	}
	if recap.Cards[1].Description == enriched.Cards[1].Description {
		t.Fatal("input recap was mutated")
	}
}

func TestApplyRejectsPartialUnknownDuplicateShareAndInvalidOutput(t *testing.T) {
	recap := sampleRecap()
	cases := []narrative.Story{
		// Empty/partial output is rejected atomically: all editable cards are required.
		{},
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: "text"}}},
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: "one"}, {ID: "unknown", Description: "two"}}},
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: "one"}, {ID: "intro", Description: "two"}}},
		// SHARE is intentionally not AI-editable.
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: "one"}, {ID: "share", Description: "two"}}},
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: strings.Repeat("я", narrative.MaxDescriptionRunes+1)}, {ID: "year-activity", Description: "two"}}},
		{Cards: []narrative.CardNarrative{{ID: "intro", Description: "скрытый\u202eтекст"}, {ID: "year-activity", Description: "two"}}},
	}
	for _, story := range cases {
		if _, err := narrative.Apply(recap, story); err == nil {
			t.Fatalf("expected validation error for %+v", story)
		}
	}
}

func TestFactsFromRecapExposeOnlyEditableCardIDsAndNoSensitiveIDs(t *testing.T) {
	recap := sampleRecap()
	facts := narrative.FactsFromRecap(recap)
	encoded, err := json.Marshal(facts)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, forbidden := range []string{
		recap.ID.String(), recap.ShareID.String(), recap.Profile.ID.String(),
		recap.ActionableState.DraftListingID.String(), recap.ActionableState.OpenDialogID.String(),
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("privacy-safe facts leaked id %s: %s", forbidden, text)
		}
	}
	if facts.Behavior.Code != model.BehaviorResearcher || facts.NextAction.Code != model.ActionOpenFavorites {
		t.Fatalf("important derived facts missing: %+v", facts)
	}
	if !reflect.DeepEqual(facts.EditableCardIDs, []string{"intro", "year-activity"}) {
		t.Fatalf("editable card ids = %#v", facts.EditableCardIDs)
	}
	if strings.Contains(text, `"share"`) {
		t.Fatalf("SHARE card id leaked into AI-editable scope: %s", text)
	}
}

func sampleRecap() model.Recap {
	profileID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	return model.Recap{
		ID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ShareID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Profile: model.Profile{ID: profileID, Code: "researcher", DisplayName: "Лена"},
		Year:    2025,
		Metrics: model.Metrics{
			TotalEvents: 200, TotalViews: 143, FavoritesAdded: 12, CategoriesCount: 7,
			TopCategoryCode: "real_estate", TopCategory: "Недвижимость", TopCategoryViews: 60,
			MostActiveMonth: 9, RepeatRate: .25, PurchaseRate: .2,
		},
		ActionableState: model.ActionableState{
			CapturedAt:     time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC),
			DraftListingID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
			OpenDialogID:   uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		},
		Behavior:     model.Behavior{Code: model.BehaviorResearcher, Title: "Глубокое исследование", Description: "Ты много сравнивал варианты."},
		Achievements: []model.Achievement{{Code: model.AchievementDecisiveStep, Title: "Решительный шаг", Reason: "Недвижимость стала заметным интересом."}},
		NextAction:   model.NextAction{Code: model.ActionOpenFavorites, Title: "Вернись к своим находкам", Description: "Открой избранное.", Reason: "Есть сохранённые объявления."},
		Cards: []model.Card{
			{ID: "intro", Type: model.CardIntro, Position: 1, Title: "Итоги", Description: "Шаблонный текст."},
			{ID: "year-activity", Type: model.CardYearActivity, Position: 2, Title: "Год в цифрах", Description: "Шаблонный текст.", Explanation: "Точные агрегаты.", Payload: model.YearActivityPayload{TotalEvents: 200, TotalViews: 143}},
			{ID: "share", Type: model.CardShare, Position: 3, Title: "Поделиться", Description: "Публичный текст должен быть детерминированным.", Shareable: true, Payload: model.ShareCard{ShareID: uuid.MustParse("33333333-3333-3333-3333-333333333333")}},
		},
	}
}

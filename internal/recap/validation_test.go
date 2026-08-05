package recap

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestValidateProfile(t *testing.T) {
	tests := []struct {
		name  string
		value Profile
		ok    bool
	}{
		{name: "valid", value: validProfile(), ok: true},
		{name: "missing id", value: Profile{Code: "x", DisplayName: "X"}},
		{name: "blank code", value: Profile{ID: testProfileID, Code: " ", DisplayName: "X"}},
		{name: "blank name", value: Profile{ID: testProfileID, Code: "x", DisplayName: "\t"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProfile(tt.value)
			if tt.ok && err != nil {
				t.Fatal(err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidProfile) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestValidateMetrics(t *testing.T) {
	valid := validMetrics()
	if err := validateMetrics(valid); err != nil {
		t.Fatal(err)
	}
	mutations := []func(*Metrics){
		func(m *Metrics) { m.TotalEvents = 1 }, func(m *Metrics) { m.UniqueListings = m.TotalViews + 1 },
		func(m *Metrics) { m.RepeatedViews++ }, func(m *Metrics) { m.ChatsWithPurchase = m.ChatsStarted + 1 },
		func(m *Metrics) { m.TopCategoryCode = "" }, func(m *Metrics) { m.MostActiveMonth = 13 },
	}
	for index, mutate := range mutations {
		m := valid
		mutate(&m)
		if err := validateMetrics(m); !errors.Is(err, ErrInvalidMetrics) {
			t.Fatalf("case %d: %v", index, err)
		}
	}
}

func TestValidateMetricsUsesExactPeriodLength(t *testing.T) {
	m := validMetrics()
	m.TotalEvents = 500
	m.ActiveDays = 366
	if err := validateMetricsForPeriod(m, RecapPeriod{Year: 2024, StartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), EndAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Final: true}); err != nil {
		t.Fatal(err)
	}
	if err := validateMetricsForPeriod(m, validPeriod()); !errors.Is(err, ErrInvalidMetrics) {
		t.Fatalf("non-leap year accepted: %v", err)
	}
}

func TestProfileUUIDJSONRoundTrip(t *testing.T) {
	original := validProfile()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Profile
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ID != original.ID {
		t.Fatalf("id changed")
	}
}

func TestValidateActionableState(t *testing.T) {
	valid := validActionableState()
	if err := validateActionableState(valid); err != nil {
		t.Fatal(err)
	}
	cases := []ActionableState{{}, {CapturedAt: fixedClock(), DraftListingID: testDraftListingID}, {CapturedAt: fixedClock(), OpenDialogID: testDialogID}, {CapturedAt: fixedClock(), ActiveListingID: testActiveListingID}}
	for _, value := range cases {
		if err := validateActionableState(value); !errors.Is(err, ErrInvalidActionableState) {
			t.Fatalf("value=%+v err=%v", value, err)
		}
	}
}

func TestValidateActionTargetOneOfAndActionCompatibility(t *testing.T) {
	valid := NextAction{Code: ActionOpenTopCategory, Title: "T", Description: "D", ButtonText: "B", Reason: "R", Target: categoryActionTarget("cars")}
	if err := validateNextAction(valid); err != nil {
		t.Fatal(err)
	}
	invalid := valid
	invalid.Target.Route = &RouteTarget{Route: "/cars"}
	if err := validateNextAction(invalid); err == nil {
		t.Fatal("multiple destinations accepted")
	}
	invalid = valid
	invalid.Target = routeActionTarget("/cars")
	if err := validateNextAction(invalid); err == nil {
		t.Fatal("wrong destination type accepted")
	}
}

func TestValidateRecap(t *testing.T) {
	if err := validateRecap(validRecap()); err != nil {
		t.Fatalf("valid recap: %v", err)
	}
	tests := []struct {
		name   string
		mutate func(*Recap)
	}{
		{name: "missing internal id", mutate: func(r *Recap) { r.ID = uuid.Nil }},
		{name: "missing share id", mutate: func(r *Recap) { r.ShareID = uuid.Nil }},
		{name: "same ids", mutate: func(r *Recap) { r.ShareID = r.ID }},
		{name: "period mismatch", mutate: func(r *Recap) { r.Period.Year = 2024 }},
		{name: "non final", mutate: func(r *Recap) { r.Period.Final = false }},
		{name: "stale rate", mutate: func(r *Recap) { r.Metrics.RepeatRate = .99 }},
		{name: "bad state", mutate: func(r *Recap) { r.ActionableState.CapturedAt = time.Time{} }},
		{name: "bad behavior score", mutate: func(r *Recap) { r.Behavior.Score++ }},
		{name: "bad action target", mutate: func(r *Recap) { r.NextAction.Target = ActionTarget{} }},
		{name: "wrong card payload", mutate: func(r *Recap) {
			r.Cards[1].Payload = TopCategoryPayload{CategoryCode: "x", Category: "X", CategoryViews: 1}
		}},
		{name: "personal card marked shareable", mutate: func(r *Recap) { r.Cards[1].Shareable = true }},
		{name: "final share card not shareable", mutate: func(r *Recap) { r.Cards[len(r.Cards)-1].Shareable = false }},
		{name: "share payload differs from public dto", mutate: func(r *Recap) {
			index := len(r.Cards) - 1
			payload := r.Cards[index].Payload.(ShareCard)
			payload.Year++
			r.Cards[index].Payload = payload
		}},
		{name: "duplicate card id", mutate: func(r *Recap) { r.Cards[1].ID = r.Cards[0].ID }},
		{name: "generated before end", mutate: func(r *Recap) { r.GeneratedAt = r.Period.StartAt; r.ActionableState.CapturedAt = r.GeneratedAt }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := validRecap()
			tt.mutate(&value)
			if err := validateRecap(value); !errors.Is(err, ErrInvalidRecap) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestValidateStoredRatesRejectsNonFiniteValues(t *testing.T) {
	m := EnrichMetrics(validMetrics())
	m.RepeatRate = math.NaN()
	if err := validateStoredRates(m); err == nil {
		t.Fatal("NaN accepted")
	}
}

func TestCardPayloadIsATypeSafeUnion(t *testing.T) {
	cards := validRecap().Cards
	for _, card := range cards {
		switch card.Type {
		case CardIntro:
			if card.Payload != nil {
				t.Fatalf("%s has payload", card.Type)
			}
		case CardShare:
			if _, ok := card.Payload.(ShareCard); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardYearActivity:
			if _, ok := card.Payload.(YearActivityPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardTopCategory:
			if _, ok := card.Payload.(TopCategoryPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardActiveMonth:
			if _, ok := card.Payload.(ActiveMonthPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardBehavior:
			if _, ok := card.Payload.(BehaviorPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardAchievement:
			if _, ok := card.Payload.(AchievementPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		case CardMissedOpportunity, CardNextAction:
			if _, ok := card.Payload.(ActionPayload); !ok {
				t.Fatalf("wrong payload %T", card.Payload)
			}
		}
	}
}

func TestValidateRecapRejectsMoreThanThreeAchievements(t *testing.T) {
	value := validRecap()
	value.Achievements = []Achievement{
		{Code: AchievementSuccessfulSeller, Category: AchievementCategorySelling, Title: "1", Description: "D", Reason: "R"},
		{Code: AchievementDealCloser, Category: AchievementCategoryBuying, Title: "2", Description: "D", Reason: "R"},
		{Code: AchievementBroadInterests, Category: AchievementCategoryDiscovery, Title: "3", Description: "D", Reason: "R"},
		{Code: AchievementMasterOfFavorites, Category: AchievementCategoryCollection, Title: "4", Description: "D", Reason: "R"},
	}
	if err := validateRecap(value); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("four achievements accepted: %v", err)
	}
}

func TestValidateRecapRejectsDuplicateAchievementCategories(t *testing.T) {
	value := validRecap()
	value.Achievements = []Achievement{
		{Code: AchievementSuccessfulSeller, Category: AchievementCategorySelling, Title: "1", Description: "D", Reason: "R"},
		{Code: AchievementConsistentPublisher, Category: AchievementCategorySelling, Title: "2", Description: "D", Reason: "R"},
	}
	if err := validateRecap(value); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("duplicate category accepted: %v", err)
	}
}

func TestValidateAchievementSelectionRejectsWrongPolicyCategoryAndOrder(t *testing.T) {
	ruleset := DefaultRuleset()
	values := ruleset.BuildAchievements(allAchievementMetrics())

	wrongCategory := append([]Achievement(nil), values...)
	wrongCategory[0].Category = AchievementCategoryCollection
	if err := validateAchievementSelection(wrongCategory, ruleset.AchievementPolicy); err == nil {
		t.Fatal("policy category mismatch accepted")
	}

	wrongOrder := append([]Achievement(nil), values...)
	wrongOrder[0], wrongOrder[1] = wrongOrder[1], wrongOrder[0]
	if err := validateAchievementSelection(wrongOrder, ruleset.AchievementPolicy); err == nil {
		t.Fatal("non-deterministic order accepted")
	}
}

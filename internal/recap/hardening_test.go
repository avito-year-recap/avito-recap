package recap

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRecapJSONRoundTripRestoresClosedCardUnion(t *testing.T) {
	original := validRecap()
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Recap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := validateRecapAgainstRuleset(decoded, DefaultRuleset(), fixedClock()); err != nil {
		t.Fatalf("round-tripped recap is invalid: %v", err)
	}
	for index := range original.Cards {
		if reflect.TypeOf(decoded.Cards[index].Payload) != reflect.TypeOf(original.Cards[index].Payload) {
			t.Fatalf("card %d payload type changed: %T -> %T", index, original.Cards[index].Payload, decoded.Cards[index].Payload)
		}
	}
}

func TestCardJSONRejectsImpossibleTypePayloadPairs(t *testing.T) {
	bad := Card{ID: "bad", Type: CardYearActivity, Position: 1, Title: "T", Description: "D", Payload: ActiveMonthPayload{Month: 1}}
	if _, err := json.Marshal(bad); err == nil {
		t.Fatal("mismatched card was serialized")
	}

	data := []byte(`{"id":"bad","type":"TOP_CATEGORY","position":1,"title":"T","description":"D","shareable":false,"payload":{"month":1}}`)
	var decoded Card
	if err := json.Unmarshal(data, &decoded); err == nil {
		t.Fatal("mismatched stored payload was accepted")
	}
}

func TestRulesDigestBindsMaterialConfiguration(t *testing.T) {
	base := DefaultRuleset()
	baseDigest := base.Digest()
	mutations := []func(*Ruleset){
		func(r *Ruleset) { r.Algorithm += "-changed" },
		func(r *Ruleset) { r.Thresholds.FindHunterMinViews++ },
		func(r *Ruleset) { r.AchievementPolicy.Rules[0].Priority++ },
		func(r *Ruleset) { r.AchievementPolicy.Rules[0].Category = AchievementCategoryBuying },
		func(r *Ruleset) { r.AchievementPolicy.MaxAwarded-- },
		func(r *Ruleset) { r.RecommendationPriorities.OpenFavorites++ },
		func(r *Ruleset) { r.SharePolicy.MaxPublicTextRunes++ },
	}
	for index, mutate := range mutations {
		value := base
		value.SharePolicy.AllowedAchievementCodes = append([]AchievementCode(nil), base.SharePolicy.AllowedAchievementCodes...)
		mutate(&value)
		if value.Digest() == baseDigest {
			t.Fatalf("mutation %d did not change digest", index)
		}
	}
}

func TestServiceRejectsPlausibleButSemanticallyForgedStoredRecap(t *testing.T) {
	stored := validRecap()
	stored.NextAction = createListingAction("Формально валидная, но не вычисленная рекомендация.")
	stored.Cards = BuildCardsWithRuleset(DefaultRuleset(), stored.Profile, stored.Year, stored.ShareID, stored.Metrics, stored.Behavior, stored.Achievements, stored.NextAction)
	if err := validateRecap(stored); err != nil {
		t.Fatalf("fixture must be structurally valid: %v", err)
	}
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored}, WithClock(fixedClock))
	if _, err := service.Get(context.Background(), stored.ID); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("forged recap error = %v", err)
	}
}

func TestServiceRejectsFutureDatedStoredRecap(t *testing.T) {
	stored := validRecap()
	stored.GeneratedAt = fixedClock().Add(24 * time.Hour)
	stored.ActionableState.CapturedAt = stored.GeneratedAt
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored}, WithClock(fixedClock))
	if _, err := service.Get(context.Background(), stored.ID); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("future recap error = %v", err)
	}
}

func TestGenerateRejectsStorageThatFalsifiesSnapshotTime(t *testing.T) {
	state := validActionableState()
	state.CapturedAt = fixedClock().Add(-time.Minute)
	service := mustServiceWithState(t,
		&profileStorageStub{profile: validProfile()}, &analyticsStorageStub{metrics: validMetrics()},
		&actionStateStorageStub{state: state}, &recapStorageStub{},
		WithClock(fixedClock), WithIDGenerator(sequenceIDGenerator(testRecapID, testShareID)),
	)
	if _, err := service.Generate(context.Background(), testProfileID, 2025); !errors.Is(err, ErrInvalidActionableState) {
		t.Fatalf("stale snapshot error = %v", err)
	}
}

func TestAddressableStateAndTargetsAreStrict(t *testing.T) {
	for _, state := range []ActionableState{
		{CapturedAt: fixedClock(), CurrentDrafts: 1},
		{CapturedAt: fixedClock(), OpenDialogs: 1},
		{CapturedAt: fixedClock(), ActiveListings: 1},
	} {
		if err := validateActionableState(state); !errors.Is(err, ErrInvalidActionableState) {
			t.Fatalf("state accepted: %+v", state)
		}
	}
	badRoute := NextAction{Code: ActionOpenFavorites, Title: "T", Description: "D", ButtonText: "B", Reason: "R", Target: routeActionTarget("//evil.example")}
	if err := validateNextAction(badRoute); err == nil {
		t.Fatal("unsafe route accepted")
	}
	badCategory := NextAction{Code: ActionOpenTopCategory, Title: "T", Description: "D", ButtonText: "B", Reason: "R", Target: categoryActionTarget("cars/../private")}
	if err := validateNextAction(badCategory); err == nil {
		t.Fatal("unsafe category code accepted")
	}
}

func TestPrivacyProjectionRejectsUnsafeExternalText(t *testing.T) {
	value := validRecap()
	value.Metrics.TopCategory = "Электроника\nprivate-token"
	value.Metrics.TopCategoryShareable = true
	card := BuildShareCardWithRuleset(DefaultRuleset(), value)
	if card.TopCategory != "" {
		t.Fatalf("unsafe category leaked: %q", card.TopCategory)
	}
	if card.PrivacyVersion != DefaultRuleset().SharePolicy.Version {
		t.Fatalf("privacy policy version missing: %+v", card)
	}
}

func TestRulesetRejectsLabelsWithoutImplementedContract(t *testing.T) {
	invalidAlgorithm := DefaultRuleset()
	invalidAlgorithm.Algorithm = "unimplemented"
	if err := invalidAlgorithm.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unknown algorithm error = %v", err)
	}

	invalidThresholds := DefaultRuleset()
	invalidThresholds.Thresholds.StartingSellerMaxPublished = invalidThresholds.Thresholds.StartingSellerMinCreated
	if err := invalidThresholds.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("impossible thresholds error = %v", err)
	}

	invalidPriority := DefaultRuleset()
	invalidPriority.RecommendationPriorities.OpenTopCategory = invalidPriority.RecommendationPriorities.FinishDraft + 1
	if err := invalidPriority.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unsafe priority error = %v", err)
	}

	invalidAchievementLimit := DefaultRuleset()
	invalidAchievementLimit.AchievementPolicy.MaxAwarded = maxAchievements + 1
	if err := invalidAchievementLimit.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("achievement limit error = %v", err)
	}

	duplicateAchievementRule := DefaultRuleset()
	duplicateAchievementRule.AchievementPolicy.Rules[1].Code = duplicateAchievementRule.AchievementPolicy.Rules[0].Code
	if err := duplicateAchievementRule.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("duplicate achievement policy error = %v", err)
	}

	unknownAchievementCategory := DefaultRuleset()
	unknownAchievementCategory.AchievementPolicy.Rules[0].Category = "UNKNOWN"
	if err := unknownAchievementCategory.Validate(); !errors.Is(err, ErrInvalidRuleset) {
		t.Fatalf("unknown achievement category error = %v", err)
	}
}

func TestServiceGetRejectsStoredAchievementFromSameCategory(t *testing.T) {
	stored := validRecap()
	stored.Achievements = []Achievement{
		{
			Code: AchievementAttentiveResearcher, Category: AchievementCategoryDiscovery,
			Title: "Внимательное сравнение", Description: "Перед следующим шагом было изучено много вариантов.",
			Reason: "Просмотров объявлений: 180.", Shareable: true,
		},
		{
			Code: AchievementBroadInterests, Category: AchievementCategoryDiscovery,
			Title: "Широкий круг интересов", Description: "Активность охватила много разных направлений.",
			Reason: "Категорий с активностью: 4.", Shareable: true,
		},
	}
	stored.Cards = BuildCardsWithRuleset(DefaultRuleset(), stored.Profile, stored.Year, stored.ShareID,
		stored.Metrics, stored.Behavior, stored.Achievements, stored.NextAction)
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored}, WithClock(fixedClock))
	if _, err := service.Get(context.Background(), stored.ID); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("same-category stored achievements error = %v", err)
	}
}

func TestServiceGetRejectsLowerGradeWhenHigherGradeWasEarned(t *testing.T) {
	stored := validRecap()
	stored.Metrics = EnrichMetrics(allAchievementMetrics())
	stored.Behavior = DefaultRuleset().DetectBehavior(stored.Metrics)
	stored.Achievements = []Achievement{
		{
			Code: AchievementConsistentPublisher, Category: AchievementCategorySelling,
			Title: "Стабильные публикации", Description: "Объявления появлялись регулярно и поддерживали активный сценарий продаж.",
			Reason: "Объявлений опубликовано: 10.", Shareable: true,
		},
		{
			Code: AchievementDealCloser, Category: AchievementCategoryBuying,
			Title: "Сделка состоялась", Description: "Выбранные варианты несколько раз превращались в завершённые покупки.",
			Reason: "Покупок завершено: 5.", Shareable: true,
		},
		{
			Code: AchievementBroadInterests, Category: AchievementCategoryDiscovery,
			Title: "Широкий круг интересов", Description: "Активность охватила много разных направлений.",
			Reason: "Категорий с активностью: 8.", Shareable: true,
		},
	}
	stored.NextAction = DefaultRuleset().BuildNextAction(stored.Metrics, stored.ActionableState, stored.Behavior)
	stored.Cards = BuildCardsWithRuleset(DefaultRuleset(), stored.Profile, stored.Year, stored.ShareID,
		stored.Metrics, stored.Behavior, stored.Achievements, stored.NextAction)
	service := mustService(t, &profileStorageStub{}, &analyticsStorageStub{}, &recapStorageStub{value: stored}, WithClock(fixedClock))
	if _, err := service.Get(context.Background(), stored.ID); !errors.Is(err, ErrInvalidRecap) {
		t.Fatalf("lower-grade stored achievement error = %v", err)
	}
}

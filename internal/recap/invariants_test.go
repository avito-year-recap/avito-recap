package recap

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRandomValidMetricsProperties(t *testing.T) {
	random := rand.New(rand.NewSource(20260805))
	for index := 0; index < 2000; index++ {
		metrics := randomValidMetrics(random)
		state := randomValidActionableState(random)
		assertPipelineInvariants(t, metrics, state)
	}
}

func TestSameRulesetAndSnapshotProduceSameResult(t *testing.T) {
	ruleset := DefaultRuleset()
	metrics := EnrichMetrics(validMetrics())
	state := validActionableState()

	firstBehavior := ruleset.DetectBehavior(metrics)
	firstAchievements := ruleset.BuildAchievements(metrics)
	firstAction := ruleset.BuildNextAction(metrics, state, firstBehavior)
	firstCards := BuildCards(validProfile(), 2025, testShareID, metrics, firstBehavior, firstAchievements, firstAction)

	secondBehavior := ruleset.DetectBehavior(metrics)
	secondAchievements := ruleset.BuildAchievements(metrics)
	secondAction := ruleset.BuildNextAction(metrics, state, secondBehavior)
	secondCards := BuildCards(validProfile(), 2025, testShareID, metrics, secondBehavior, secondAchievements, secondAction)

	if !reflect.DeepEqual(firstBehavior, secondBehavior) {
		t.Fatalf("behavior is not deterministic:\nfirst:  %+v\nsecond: %+v", firstBehavior, secondBehavior)
	}
	if !reflect.DeepEqual(firstAchievements, secondAchievements) {
		t.Fatalf("achievements are not deterministic:\nfirst:  %+v\nsecond: %+v", firstAchievements, secondAchievements)
	}
	if !reflect.DeepEqual(firstAction, secondAction) {
		t.Fatalf("action is not deterministic:\nfirst:  %+v\nsecond: %+v", firstAction, secondAction)
	}
	if !reflect.DeepEqual(firstCards, secondCards) {
		t.Fatalf("cards are not deterministic:\nfirst:  %+v\nsecond: %+v", firstCards, secondCards)
	}
}

func TestGeneratedUserFacingTextHasNoBoundaryWhitespace(t *testing.T) {
	profile := validProfile()
	profile.Code = "\tactive-buyer  "
	profile.DisplayName = "  Алексей\n"
	profile.Description = "  Тестовый профиль\t"
	profile.AvatarURL = "  /avatars/alexey.png  "

	metrics := validMetrics()
	metrics.TopCategoryCode = "\telectronics "
	metrics.TopCategory = "  Электроника\n"
	metrics = EnrichMetrics(metrics)
	state := validActionableState()
	ruleset := DefaultRuleset()
	behavior := ruleset.DetectBehavior(metrics)
	achievements := ruleset.BuildAchievements(metrics)
	action := ruleset.BuildNextAction(metrics, state, behavior)
	cards := BuildCards(profile, 2025, testShareID, metrics, behavior, achievements, action)

	value := Recap{
		ID:              testRecapID,
		ShareID:         testShareID,
		Profile:         profile,
		Year:            2025,
		Period:          validPeriod(),
		RulesVersion:    "  " + ruleset.Version + "\n",
		Metrics:         metrics,
		ActionableState: state,
		Behavior:        behavior,
		Achievements:    achievements,
		Cards:           cards,
		NextAction:      action,
		GeneratedAt:     fixedClock(),
	}
	value = normalizeRecap(value)

	assertRecapStringsNormalized(t, value)
	if err := validateRecap(value); err != nil {
		t.Fatalf("normalized recap must remain valid: %v", err)
	}
}

func assertPipelineInvariants(t testing.TB, metrics Metrics, state ActionableState) {
	t.Helper()
	if err := validateMetricsForPeriod(metrics, validPeriod()); err != nil {
		t.Fatalf("generated metrics are invalid: %+v: %v", metrics, err)
	}
	assertRateInUnitInterval(t, "repeat rate", metrics.RepeatRate)
	assertRateInUnitInterval(t, "purchase rate", metrics.PurchaseRate)

	ruleset := DefaultRuleset()
	behavior := ruleset.DetectBehavior(metrics)
	achievements := ruleset.BuildAchievements(metrics)
	action := ruleset.BuildNextAction(metrics, state, behavior)
	if err := validateNextAction(action); err != nil {
		t.Fatalf("action does not contain a valid required target: %+v: %v", action, err)
	}
	cards := BuildCards(validProfile(), 2025, testShareID, metrics, behavior, achievements, action)
	if err := validateCards(cards); err != nil {
		t.Fatalf("cards violate identity/position/payload invariants: %v", err)
	}

	value := Recap{
		ID:              testRecapID,
		ShareID:         testShareID,
		Profile:         validProfile(),
		Year:            2025,
		Period:          validPeriod(),
		RulesVersion:    ruleset.Version,
		Metrics:         metrics,
		ActionableState: state,
		Behavior:        behavior,
		Achievements:    achievements,
		Cards:           cards,
		NextAction:      action,
		GeneratedAt:     fixedClock(),
	}
	if err := validateRecap(value); err != nil {
		t.Fatalf("full recap pipeline produced invalid data: %+v: %v", value, err)
	}
	assertRecapStringsNormalized(t, value)
}

func assertRateInUnitInterval(t testing.TB, name string, value float64) {
	t.Helper()
	if value < 0 || value > 1 {
		t.Fatalf("%s = %v, want value in [0,1]", name, value)
	}
}

func assertRecapStringsNormalized(t testing.TB, value Recap) {
	t.Helper()
	check := func(path, text string, required bool) {
		t.Helper()
		if text != strings.TrimSpace(text) {
			t.Fatalf("%s contains boundary whitespace: %q", path, text)
		}
		if required && text == "" {
			t.Fatalf("%s is unexpectedly empty", path)
		}
	}

	check("profile.code", value.Profile.Code, true)
	check("profile.displayName", value.Profile.DisplayName, true)
	check("profile.description", value.Profile.Description, false)
	check("profile.avatarUrl", value.Profile.AvatarURL, false)
	check("rulesVersion", value.RulesVersion, true)
	check("metrics.topCategoryCode", value.Metrics.TopCategoryCode, value.Metrics.TopCategory != "")
	check("metrics.topCategory", value.Metrics.TopCategory, value.Metrics.TopCategoryCode != "")
	check("behavior.title", value.Behavior.Title, true)
	check("behavior.description", value.Behavior.Description, true)
	check("behavior.reason", value.Behavior.Reason, true)
	for index, evidence := range value.Behavior.Evidence {
		check(fmt.Sprintf("behavior.evidence[%d].metric", index), evidence.Metric, true)
		check(fmt.Sprintf("behavior.evidence[%d].detail", index), evidence.Detail, true)
	}
	for index, achievement := range value.Achievements {
		check(fmt.Sprintf("achievements[%d].title", index), achievement.Title, true)
		check(fmt.Sprintf("achievements[%d].description", index), achievement.Description, true)
		check(fmt.Sprintf("achievements[%d].reason", index), achievement.Reason, true)
	}
	check("nextAction.title", value.NextAction.Title, true)
	check("nextAction.description", value.NextAction.Description, true)
	check("nextAction.buttonText", value.NextAction.ButtonText, true)
	check("nextAction.reason", value.NextAction.Reason, true)
	assertTargetStringsNormalized(t, "nextAction.target", value.NextAction.Target)
	for index, card := range value.Cards {
		prefix := fmt.Sprintf("cards[%d]", index)
		check(prefix+".id", card.ID, true)
		check(prefix+".title", card.Title, true)
		check(prefix+".description", card.Description, true)
		check(prefix+".explanation", card.Explanation, false)
		switch payload := card.Payload.(type) {
		case TopCategoryPayload:
			check(prefix+".payload.categoryCode", payload.CategoryCode, true)
			check(prefix+".payload.category", payload.Category, true)
		case BehaviorPayload:
			for evidenceIndex, evidence := range payload.Evidence {
				check(fmt.Sprintf("%s.payload.evidence[%d].metric", prefix, evidenceIndex), evidence.Metric, true)
				check(fmt.Sprintf("%s.payload.evidence[%d].detail", prefix, evidenceIndex), evidence.Detail, true)
			}
		case ActionPayload:
			assertTargetStringsNormalized(t, prefix+".payload.target", payload.Target)
		case ShareCard:
			check(prefix+".payload.behaviorTitle", payload.BehaviorTitle, true)
			check(prefix+".payload.achievementTitle", payload.AchievementTitle, false)
			check(prefix+".payload.topCategory", payload.TopCategory, false)
		}
	}
}

func assertTargetStringsNormalized(t testing.TB, path string, target ActionTarget) {
	t.Helper()
	if target.Route != nil && target.Route.Route != strings.TrimSpace(target.Route.Route) {
		t.Fatalf("%s.route contains boundary whitespace: %q", path, target.Route.Route)
	}
	if target.Category != nil && target.Category.CategoryCode != strings.TrimSpace(target.Category.CategoryCode) {
		t.Fatalf("%s.category contains boundary whitespace: %q", path, target.Category.CategoryCode)
	}
	if target.Search != nil && target.Search.CategoryCode != strings.TrimSpace(target.Search.CategoryCode) {
		t.Fatalf("%s.search contains boundary whitespace: %q", path, target.Search.CategoryCode)
	}
}

func randomValidMetrics(random *rand.Rand) Metrics {
	searches := uint64(random.Intn(80))
	views := uint64(random.Intn(250))
	favorites := uint64(random.Intn(60))
	chats := uint64(random.Intn(35))
	created := uint64(random.Intn(25))
	published := uint64(random.Intn(25))
	purchases := uint64(random.Intn(15))
	sales := uint64(random.Intn(15))
	total := searches + views + favorites + chats + created + published + purchases + sales
	if total < minEventsForRecap {
		searches += minEventsForRecap - total
		total = minEventsForRecap
	}

	unique := boundedRandom(random, views)
	chatsWithPurchase := boundedRandom(random, chats)
	activeDays := boundedRandom(random, minUint64(total, 365))

	metrics := Metrics{
		TotalEvents:        total,
		Searches:           searches,
		TotalViews:         views,
		UniqueListings:     unique,
		RepeatedViews:      views - unique,
		FavoritesAdded:     favorites,
		ChatsStarted:       chats,
		ListingsCreated:    created,
		ListingsPublished:  published,
		PurchasesCompleted: purchases,
		SalesCompleted:     sales,
		ActiveDays:         activeDays,
		ChatsWithPurchase:  chatsWithPurchase,
		MostActiveMonth:    uint32(random.Intn(13)),
	}
	if views > 0 && random.Intn(2) == 1 {
		metrics.CategoriesCount = 1 + uint64(random.Intn(int(minUint64(total, 10))))
		metrics.TopCategoryCode = "category"
		metrics.TopCategory = "Категория"
		metrics.TopCategoryViews = 1 + uint64(random.Intn(int(views)))
		metrics.TopCategoryShareable = random.Intn(2) == 1
	}
	return EnrichMetrics(metrics)
}

func randomValidActionableState(random *rand.Rand) ActionableState {
	state := ActionableState{
		CapturedAt:                   fixedClock(),
		FavoritesCount:               uint64(random.Intn(20)),
		HasSavedSearchForTopCategory: random.Intn(2) == 1,
		HasEverPublishedListing:      random.Intn(2) == 1,
	}
	if random.Intn(4) == 0 {
		state.CurrentDrafts = 1 + uint64(random.Intn(5))
		state.DraftListingID = testDraftListingID
	}
	if random.Intn(4) == 0 {
		state.OpenDialogs = 1 + uint64(random.Intn(5))
		state.OpenDialogID = testDialogID
	}
	if random.Intn(4) == 0 {
		state.ActiveListings = 1 + uint64(random.Intn(5))
		state.ActiveListingID = testActiveListingID
	}
	if random.Intn(4) == 0 {
		state.LastPurchasedListingID = uuid.MustParse("66666666-6666-4666-8666-666666666666")
	}
	return state
}

func boundedRandom(random *rand.Rand, maximum uint64) uint64 {
	if maximum == 0 {
		return 0
	}
	return uint64(random.Int63n(int64(maximum) + 1))
}

func minUint64(left, right uint64) uint64 {
	if left < right {
		return left
	}
	return right
}

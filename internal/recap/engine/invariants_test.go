package engine_test

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/engine"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/testkit"
	"github.com/year-recap/internal/recap/validation/structural"
)

func TestRandomValidMetricsProperties(t *testing.T) {
	random := rand.New(rand.NewSource(20260805))
	for index := 0; index < 2000; index++ {
		assertPipelineInvariants(t, randomValidMetrics(random), randomValidActionableState(random))
	}
}

func TestSameRulesetAndSnapshotProduceSameResult(t *testing.T) {
	core := testEngine(t)
	build := func() model.Recap {
		value, err := core.Build(engine.BuildInput{
			RecapID: testkit.RecapID, ShareID: testkit.ShareID, Profile: testkit.Profile(), Year: 2025,
			Period: testkit.Period(), Metrics: testkit.Metrics(), ActionableState: testkit.ActionableState(), GeneratedAt: testkit.Clock(),
		})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		return value
	}
	first, second := build(), build()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same ruleset and snapshot produced different derived data")
	}
}

func TestBuildNormalizesBoundaryStringsOnce(t *testing.T) {
	profile := testkit.Profile()
	profile.Code = "\tactive-buyer  "
	profile.DisplayName = "  Алексей\n"
	profile.Description = "  Тестовый профиль\t"
	profile.AvatarURL = "  /avatars/alexey.png  "
	metrics := testkit.Metrics()
	metrics.TopCategoryCode = "\telectronics "
	metrics.TopCategory = "  Электроника\n"
	value, err := testEngine(t).Build(engine.BuildInput{
		RecapID: testkit.RecapID, ShareID: testkit.ShareID, Profile: profile, Year: 2025,
		Period: testkit.Period(), Metrics: metrics, ActionableState: testkit.ActionableState(), GeneratedAt: testkit.Clock(),
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	assertRecapStringsNormalized(t, value)
	if err := structural.ValidateRecap(value); err != nil {
		t.Fatalf("normalized recap must remain valid: %v", err)
	}
}

func assertPipelineInvariants(t testing.TB, metrics model.Metrics, state model.ActionableState) {
	t.Helper()
	core := testEngine(t)
	value, err := core.Build(engine.BuildInput{
		RecapID: testkit.RecapID, ShareID: testkit.ShareID, Profile: testkit.Profile(), Year: 2025,
		Period: testkit.Period(), Metrics: metrics, ActionableState: state, GeneratedAt: testkit.Clock(),
	})
	if err != nil {
		t.Fatalf("engine rejected generated input: %+v: %v", metrics, err)
	}
	assertRateInUnitInterval(t, "repeat rate", value.Metrics.RepeatRate)
	assertRateInUnitInterval(t, "purchase rate", value.Metrics.PurchaseRate)
	if err := structural.ValidateCards(value.Cards); err != nil {
		t.Fatalf("cards violate invariants: %v", err)
	}
	if _, err := core.ValidateStored(value, testkit.Clock()); err != nil {
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

func assertRecapStringsNormalized(t testing.TB, value model.Recap) {
	t.Helper()
	check := func(path, text string, required bool) {
		if text != strings.TrimSpace(text) {
			t.Fatalf("%s contains boundary whitespace: %q", path, text)
		}
		if required && text == "" {
			t.Fatalf("%s is unexpectedly empty", path)
		}
	}
	check("profile.code", value.Profile.Code, true)
	check("profile.displayName", value.Profile.DisplayName, true)
	check("rulesVersion", value.RulesVersion, true)
	check("behavior.title", value.Behavior.Title, true)
	check("behavior.description", value.Behavior.Description, true)
	check("behavior.reason", value.Behavior.Reason, true)
	for index, item := range value.Behavior.Evidence {
		check(fmt.Sprintf("behavior.evidence[%d].metric", index), item.Metric, true)
		check(fmt.Sprintf("behavior.evidence[%d].detail", index), item.Detail, true)
	}
	for index, item := range value.Achievements {
		check(fmt.Sprintf("achievements[%d].title", index), item.Title, true)
		check(fmt.Sprintf("achievements[%d].description", index), item.Description, true)
		check(fmt.Sprintf("achievements[%d].reason", index), item.Reason, true)
	}
	check("nextAction.title", value.NextAction.Title, true)
	check("nextAction.description", value.NextAction.Description, true)
	check("nextAction.buttonText", value.NextAction.ButtonText, true)
	check("nextAction.reason", value.NextAction.Reason, true)
	for index, card := range value.Cards {
		check(fmt.Sprintf("cards[%d].id", index), card.ID, true)
		check(fmt.Sprintf("cards[%d].title", index), card.Title, true)
		check(fmt.Sprintf("cards[%d].description", index), card.Description, true)
	}
}

func randomValidMetrics(random *rand.Rand) model.Metrics {
	searches := uint64(random.Intn(80))
	views := uint64(random.Intn(250))
	favorites := uint64(random.Intn(60))
	chats := uint64(random.Intn(35))
	created := uint64(random.Intn(25))
	published := uint64(random.Intn(25))
	purchases := uint64(random.Intn(15))
	sales := uint64(random.Intn(15))
	total := searches + views + favorites + chats + created + published + purchases + sales
	if total < 5 {
		searches += 5 - total
		total = 5
	}
	unique := boundedRandom(random, views)
	metrics := model.Metrics{
		TotalEvents: total, Searches: searches, TotalViews: views, UniqueListings: unique, RepeatedViews: views - unique,
		FavoritesAdded: favorites, ChatsStarted: chats, ListingsCreated: created, ListingsPublished: published,
		PurchasesCompleted: purchases, SalesCompleted: sales, ActiveDays: boundedRandom(random, minUint64(total, 365)),
		ChatsWithPurchase: boundedRandom(random, chats), MostActiveMonth: uint32(1 + random.Intn(12)),
	}
	if views > 0 && random.Intn(2) == 1 {
		metrics.CategoriesCount = 1 + uint64(random.Intn(int(minUint64(total, 10))))
		metrics.TopCategoryCode = "category"
		metrics.TopCategory = "Категория"
		metrics.TopCategoryViews = 1 + uint64(random.Intn(int(views)))
		metrics.TopCategoryShareable = random.Intn(2) == 1
	}
	return metrics
}

func randomValidActionableState(random *rand.Rand) model.ActionableState {
	state := model.ActionableState{
		CapturedAt: testkit.Clock(), FavoritesCount: uint64(random.Intn(20)),
		HasSavedSearchForTopCategory: random.Intn(2) == 1, HasEverPublishedListing: random.Intn(2) == 1,
	}
	if random.Intn(4) == 0 {
		state.CurrentDrafts = 1 + uint64(random.Intn(5))
		state.DraftListingID = testkit.DraftListingID
	}
	if random.Intn(4) == 0 {
		state.OpenDialogs = 1 + uint64(random.Intn(5))
		state.OpenDialogID = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	}
	if random.Intn(4) == 0 {
		state.ActiveListings = 1 + uint64(random.Intn(5))
		state.ActiveListingID = uuid.MustParse("44444444-4444-4444-8444-444444444444")
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

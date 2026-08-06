package recap

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type expectedRecap struct {
	ProfileCode              string            `json:"profileCode"`
	ExpectedRulesVersion     string            `json:"expectedRulesVersion"`
	ExpectedBehavior         BehaviorCode      `json:"expectedBehavior"`
	ExpectedAchievements     []AchievementCode `json:"expectedAchievements"`
	ExpectedNextAction       ActionCode        `json:"expectedNextAction"`
	ExpectedTopCategoryCode  string            `json:"expectedTopCategoryCode"`
	ExpectedTopCategoryTitle string            `json:"expectedTopCategoryTitle"`
	ExpectedActiveMonth      uint32            `json:"expectedActiveMonth"`
	RequiredCards            []CardType        `json:"requiredCards"`
}

type behaviorCase struct {
	Name             string       `json:"name"`
	Metrics          Metrics      `json:"metrics"`
	ExpectedBehavior BehaviorCode `json:"expectedBehavior"`
}

func TestSeedProfilesGenerateExpectedRecaps(t *testing.T) {
	var profiles []Profile
	var scenarios []SeedScenario
	var expected []expectedRecap
	readJSONFile(t, projectFile(t, "seeds", "profiles.json"), &profiles)
	readJSONFile(t, projectFile(t, "seeds", "scenarios.json"), &scenarios)
	readJSONFile(t, projectFile(t, "testdata", "expected", "recaps.json"), &expected)

	seenUUIDs := make(map[uuid.UUID]string)
	registerUUID := func(id uuid.UUID, owner string) {
		if id == uuid.Nil {
			return
		}
		if previous, exists := seenUUIDs[id]; exists {
			t.Fatalf("seed UUID %s is reused by %s and %s", id, previous, owner)
		}
		seenUUIDs[id] = owner
	}

	profilesByCode := make(map[string]Profile, len(profiles))
	for _, profile := range profiles {
		if err := validateProfile(profile); err != nil {
			t.Fatalf("invalid seed profile %q: %v", profile.Code, err)
		}
		if _, exists := profilesByCode[profile.Code]; exists {
			t.Fatalf("duplicate profile code %q", profile.Code)
		}
		registerUUID(profile.ID, "profile:"+profile.Code)
		if !strings.HasPrefix(profile.AvatarURL, "/avatars/") || !strings.HasSuffix(profile.AvatarURL, ".png") {
			t.Fatalf("profile %q avatar must be a local PNG placeholder, got %q", profile.Code, profile.AvatarURL)
		}
		avatarPath := projectFile(t, "static", filepath.FromSlash(strings.TrimPrefix(profile.AvatarURL, "/")))
		if info, err := os.Stat(avatarPath); err != nil || info.IsDir() {
			t.Fatalf("profile %q avatar placeholder is unavailable at %s: %v", profile.Code, avatarPath, err)
		}
		profilesByCode[profile.Code] = profile
	}

	scenariosByCode := make(map[string]SeedScenario, len(scenarios))
	for _, scenario := range scenarios {
		if _, exists := scenariosByCode[scenario.ProfileCode]; exists {
			t.Fatalf("duplicate scenario for %q", scenario.ProfileCode)
		}
		registerUUID(scenario.ActionableState.DraftListingID, "draftListing:"+scenario.ProfileCode)
		registerUUID(scenario.ActionableState.OpenDialogID, "openDialog:"+scenario.ProfileCode)
		registerUUID(scenario.ActionableState.ActiveListingID, "activeListing:"+scenario.ProfileCode)
		registerUUID(scenario.ActionableState.LastPurchasedListingID, "lastPurchasedListing:"+scenario.ProfileCode)
		scenariosByCode[scenario.ProfileCode] = scenario
	}

	if len(profiles) != len(scenarios) || len(profiles) != len(expected) {
		t.Fatalf("profiles/scenarios/expected count mismatch: %d/%d/%d", len(profiles), len(scenarios), len(expected))
	}

	for _, want := range expected {
		want := want
		t.Run(want.ProfileCode, func(t *testing.T) {
			profile, ok := profilesByCode[want.ProfileCode]
			if !ok {
				t.Fatalf("profile %q not found", want.ProfileCode)
			}
			scenario, ok := scenariosByCode[want.ProfileCode]
			if !ok {
				t.Fatalf("scenario %q not found", want.ProfileCode)
			}

			metrics, err := MetricsFromScenario(scenario)
			if err != nil {
				t.Fatalf("build metrics: %v", err)
			}
			storage := &recapStorageStub{}
			service := mustServiceWithState(
				t,
				&profileStorageStub{profile: profile},
				&analyticsStorageStub{metrics: metrics},
				&actionStateStorageStub{state: scenario.ActionableState},
				storage,
				WithClock(fixedClock),
				WithIDGenerator(func() func() (uuid.UUID, error) {
					call := byte(0)
					return func() (uuid.UUID, error) {
						value := profile.ID
						value[0] ^= 0xff
						value[1] ^= call
						call++
						return value, nil
					}
				}()),
			)

			actual, err := service.Generate(context.Background(), profile.ID, scenario.Year)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if actual.RulesVersion != want.ExpectedRulesVersion || actual.RulesVersion != CurrentRulesVersion {
				t.Fatalf("rules version = %q, want %q", actual.RulesVersion, want.ExpectedRulesVersion)
			}
			if actual.Behavior.Code != want.ExpectedBehavior {
				t.Fatalf("behavior = %s, want %s; reason: %s", actual.Behavior.Code, want.ExpectedBehavior, actual.Behavior.Reason)
			}
			assertAchievementCodes(t, actual.Achievements, want.ExpectedAchievements)
			if actual.NextAction.Code != want.ExpectedNextAction {
				t.Fatalf("next action = %s, want %s; reason: %s", actual.NextAction.Code, want.ExpectedNextAction, actual.NextAction.Reason)
			}
			if actual.Metrics.TopCategoryCode != want.ExpectedTopCategoryCode ||
				actual.Metrics.TopCategory != want.ExpectedTopCategoryTitle {
				t.Fatalf("top category = %q/%q, want %q/%q", actual.Metrics.TopCategoryCode, actual.Metrics.TopCategory, want.ExpectedTopCategoryCode, want.ExpectedTopCategoryTitle)
			}
			if actual.Metrics.MostActiveMonth != want.ExpectedActiveMonth {
				t.Fatalf("active month = %d, want %d", actual.Metrics.MostActiveMonth, want.ExpectedActiveMonth)
			}
			assertCardSequence(t, actual.Cards, want.RequiredCards)
			assertGoldenRecap(t, actual, projectFile(t, "testdata", "golden", want.ProfileCode+".json"))
			if storage.saved == nil {
				t.Fatal("generated recap was not saved")
			}
		})
	}
}

func TestSeedCatalogueCoversAllMVPOutcomes(t *testing.T) {
	var expected []expectedRecap
	readJSONFile(t, projectFile(t, "testdata", "expected", "recaps.json"), &expected)

	behaviors := make(map[BehaviorCode]bool)
	achievements := make(map[AchievementCode]bool)
	actions := make(map[ActionCode]bool)
	for _, recap := range expected {
		behaviors[recap.ExpectedBehavior] = true
		actions[recap.ExpectedNextAction] = true
		for _, code := range recap.ExpectedAchievements {
			achievements[code] = true
		}
	}

	for _, code := range []BehaviorCode{
		BehaviorActiveSeller, BehaviorStartingSeller, BehaviorDecisiveBuyer,
		BehaviorFindHunter, BehaviorResearcher, BehaviorUniversal,
	} {
		if !behaviors[code] {
			t.Errorf("seed catalogue does not cover behavior %s", code)
		}
	}
	for _, code := range []AchievementCode{
		AchievementSuccessfulSeller, AchievementConsistentPublisher,
		AchievementAttentiveResearcher, AchievementMasterOfFavorites,
		AchievementBroadInterests, AchievementAllRounder,
		AchievementFirstSellingSteps, AchievementDealCloser,
		AchievementQuickDecision,
	} {
		if !achievements[code] {
			t.Errorf("seed catalogue does not cover achievement %s", code)
		}
	}
	for _, code := range []ActionCode{
		ActionFinishDraft, ActionOpenFavorites, ActionOpenTopCategory, ActionExploreRecommendations,
	} {
		if !actions[code] {
			t.Errorf("seed catalogue does not cover action %s", code)
		}
	}
}

func TestSellerBuyerHybridSeedQualifiesForBothBehaviorRules(t *testing.T) {
	var scenarios []SeedScenario
	readJSONFile(t, projectFile(t, "seeds", "scenarios.json"), &scenarios)

	var hybrid SeedScenario
	found := false
	for _, scenario := range scenarios {
		if scenario.ProfileCode == "seller-buyer-hybrid" {
			hybrid = scenario
			found = true
			break
		}
	}
	if !found {
		t.Fatal("seller-buyer-hybrid seed not found")
	}

	metrics, err := MetricsFromScenario(hybrid)
	if err != nil {
		t.Fatalf("build hybrid metrics: %v", err)
	}
	thresholds := DefaultRuleset().Thresholds
	if !evaluateActiveSeller(metrics, thresholds).eligible {
		t.Fatal("hybrid seed does not qualify for ACTIVE_SELLER")
	}
	if !evaluateDecisiveBuyer(metrics, thresholds).eligible {
		t.Fatal("hybrid seed does not qualify for DECISIVE_BUYER")
	}
	if actual := DetectBehavior(metrics).Code; actual != BehaviorActiveSeller {
		t.Fatalf("hybrid tie-break behavior = %s, want %s", actual, BehaviorActiveSeller)
	}
}

func TestBehaviorCasesJSONMatchesRules(t *testing.T) {
	var cases []behaviorCase
	readJSONFile(t, projectFile(t, "testdata", "metrics", "behavior_cases.json"), &cases)
	if len(cases) == 0 {
		t.Fatal("behavior cases are empty")
	}
	for _, test := range cases {
		t.Run(test.Name, func(t *testing.T) {
			if actual := DetectBehavior(test.Metrics).Code; actual != test.ExpectedBehavior {
				t.Fatalf("behavior = %s, want %s", actual, test.ExpectedBehavior)
			}
		})
	}
}

func assertGoldenRecap(t *testing.T, actual Recap, path string) {
	t.Helper()
	data, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("marshal full golden recap: %v", err)
	}
	data = append(data, '\n')

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("update %s: %v", path, err)
		}
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !bytes.Equal(expected, data) {
		t.Fatalf("full golden recap mismatch for %s\n--- expected ---\n%s\n--- actual ---\n%s\nRun UPDATE_GOLDEN=1 go test ./internal/recap -run TestSeedProfilesGenerateExpectedRecaps after reviewing intentional changes.", path, expected, data)
	}
}

func projectFile(t *testing.T, parts ...string) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
	all := append([]string{root}, parts...)
	return filepath.Join(all...)
}

func readJSONFile(t *testing.T, path string, target any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

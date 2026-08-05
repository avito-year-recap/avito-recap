package recap

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

type goldenRecap struct {
	Profile struct {
		Code        string `json:"code"`
		DisplayName string `json:"displayName"`
	} `json:"profile"`
	Year         uint32 `json:"year"`
	RulesVersion string `json:"rulesVersion"`
	Metrics      struct {
		TopCategoryCode string `json:"topCategoryCode"`
		TopCategory     string `json:"topCategory"`
		MostActiveMonth uint32 `json:"mostActiveMonth"`
	} `json:"metrics"`
	Behavior struct {
		Code BehaviorCode `json:"code"`
	} `json:"behavior"`
	Achievements []struct {
		Code      AchievementCode `json:"code"`
		Shareable bool            `json:"shareable"`
	} `json:"achievements"`
	NextAction struct {
		Code ActionCode `json:"code"`
	} `json:"nextAction"`
	Cards []struct {
		Type     CardType `json:"type"`
		Position uint32   `json:"position"`
	} `json:"cards"`
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

	profilesByCode := make(map[string]Profile, len(profiles))
	for _, profile := range profiles {
		if err := validateProfile(profile); err != nil {
			t.Fatalf("invalid seed profile %q: %v", profile.Code, err)
		}
		if _, exists := profilesByCode[profile.Code]; exists {
			t.Fatalf("duplicate profile code %q", profile.Code)
		}
		profilesByCode[profile.Code] = profile
	}

	scenariosByCode := make(map[string]SeedScenario, len(scenarios))
	for _, scenario := range scenarios {
		if _, exists := scenariosByCode[scenario.ProfileCode]; exists {
			t.Fatalf("duplicate scenario for %q", scenario.ProfileCode)
		}
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
			service := mustService(
				t,
				&profileStorageStub{profile: profile},
				&analyticsStorageStub{metrics: metrics},
				storage,
				WithClock(fixedClock),
				WithIDGenerator(func() (string, error) { return profile.ID, nil }),
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
	var golden goldenRecap
	readJSONFile(t, path, &golden)

	if golden.Profile.Code != actual.Profile.Code || golden.Profile.DisplayName != actual.Profile.DisplayName {
		t.Fatalf("golden profile mismatch: %+v vs %+v", golden.Profile, actual.Profile)
	}
	if golden.Year != actual.Year || golden.RulesVersion != actual.RulesVersion {
		t.Fatalf("golden metadata mismatch: year=%d version=%q", golden.Year, golden.RulesVersion)
	}
	if golden.Metrics.TopCategoryCode != actual.Metrics.TopCategoryCode ||
		golden.Metrics.TopCategory != actual.Metrics.TopCategory ||
		golden.Metrics.MostActiveMonth != actual.Metrics.MostActiveMonth {
		t.Fatalf("golden metrics mismatch: %+v vs %+v", golden.Metrics, actual.Metrics)
	}
	if golden.Behavior.Code != actual.Behavior.Code || golden.NextAction.Code != actual.NextAction.Code {
		t.Fatalf("golden behavior/action mismatch: %s/%s vs %s/%s", golden.Behavior.Code, golden.NextAction.Code, actual.Behavior.Code, actual.NextAction.Code)
	}
	if len(golden.Achievements) != len(actual.Achievements) {
		t.Fatalf("golden achievement count = %d, actual = %d", len(golden.Achievements), len(actual.Achievements))
	}
	for index, achievement := range golden.Achievements {
		if achievement.Code != actual.Achievements[index].Code || achievement.Shareable != actual.Achievements[index].Shareable {
			t.Fatalf("golden achievement %d mismatch: %+v vs %+v", index, achievement, actual.Achievements[index])
		}
	}
	if len(golden.Cards) != len(actual.Cards) {
		t.Fatalf("golden card count = %d, actual = %d", len(golden.Cards), len(actual.Cards))
	}
	for index, card := range golden.Cards {
		if card.Type != actual.Cards[index].Type || card.Position != actual.Cards[index].Position {
			t.Fatalf("golden card %d mismatch: %+v vs %+v", index, card, actual.Cards[index])
		}
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

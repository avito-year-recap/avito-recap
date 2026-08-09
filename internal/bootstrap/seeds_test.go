package bootstrap

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/year-recap/internal/recap/testkit"
)

func TestMetricsFromScenarioMatchesDomainFixtureContract(t *testing.T) {
	scenarios, err := readJSON[[]scenario]("../../seeds/scenarios.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range scenarios {
		t.Run(item.ProfileCode, func(t *testing.T) {
			got, err := metricsFromScenario(item)
			if err != nil {
				t.Fatal(err)
			}

			raw, err := json.Marshal(item)
			if err != nil {
				t.Fatal(err)
			}
			var fixture testkit.SeedScenario
			if err := json.Unmarshal(raw, &fixture); err != nil {
				t.Fatal(err)
			}
			want, err := testkit.MetricsFromScenario(fixture)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("bootstrap metrics drifted from domain fixture contract:\n got: %+v\nwant: %+v", got, want)
			}
		})
	}
}

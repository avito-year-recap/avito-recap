package testkit

import (
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/seed"
)

var ErrInvalidScenario = seed.ErrInvalidScenario

type SeedScenario = seed.Scenario
type SeedActivity = seed.Activity
type WeightedCategory = seed.WeightedCategory
type WeightedMonth = seed.WeightedMonth

func MetricsFromScenario(scenario SeedScenario) (model.Metrics, error) {
	return seed.MetricsFromScenario(scenario)
}

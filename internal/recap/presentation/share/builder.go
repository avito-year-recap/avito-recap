package share

import (
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

func BuildWithRuleset(configured ruleset.Ruleset, value model.Recap) model.ShareCard {
	return Build(
		configured.SharePolicy,
		value.ShareID,
		value.Year,
		value.Metrics,
		value.Behavior,
		value.Achievements,
	)
}

package engine

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
)

var ErrNotEnoughActivity = errors.New("not enough activity to generate recap")

type Engine struct {
	rules  ruleset.Ruleset
	digest string
}

func New(configured ruleset.Ruleset) (*Engine, error) {
	configured.Version = model.NormalizeString(configured.Version)
	configured.Algorithm = model.NormalizeString(configured.Algorithm)
	configured.SharePolicy.Version = model.NormalizeString(configured.SharePolicy.Version)
	configured.AchievementPolicy.Rules = append([]ruleset.AchievementRuleConfig(nil), configured.AchievementPolicy.Rules...)
	configured.SharePolicy.AllowedAchievementCodes = append([]model.AchievementCode(nil), configured.SharePolicy.AllowedAchievementCodes...)
	if err := configured.Validate(); err != nil {
		return nil, err
	}
	return &Engine{rules: configured, digest: configured.Digest()}, nil
}

func (e *Engine) RecapKey(profileID uuid.UUID, year uint32) model.RecapKey {
	return model.RecapKey{
		ProfileID:    profileID,
		Year:         year,
		RulesVersion: e.rules.Version,
		RulesDigest:  e.digest,
	}
}

func (e *Engine) ensureEligible(metrics model.Metrics) error {
	if metrics.TotalEvents < e.rules.Eligibility.MinEvents {
		return fmt.Errorf("%w: got %d events, minimum is %d", ErrNotEnoughActivity, metrics.TotalEvents, e.rules.Eligibility.MinEvents)
	}
	return nil
}

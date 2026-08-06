package structural

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"math"
	"strings"
	"time"
)

func ValidateRecap(value model.Recap) error {
	if value.ID == uuid.Nil {
		return fmt.Errorf("%w: internal id is required", ErrInvalidRecap)
	}
	if value.ShareID == uuid.Nil {
		return fmt.Errorf("%w: public share id is required", ErrInvalidRecap)
	}
	if value.ID == value.ShareID {
		return fmt.Errorf("%w: internal and public ids must differ", ErrInvalidRecap)
	}
	if err := ValidateProfile(value.Profile); err != nil {
		return fmt.Errorf("%w: profile: %v", ErrInvalidRecap, err)
	}
	if value.Year == 0 || value.Period.Year != value.Year {
		return fmt.Errorf("%w: year and period are inconsistent", ErrInvalidRecap)
	}
	if err := ValidatePeriod(value.Period); err != nil {
		return fmt.Errorf("%w: period: %v", ErrInvalidRecap, err)
	}
	if !semanticVersionPattern.MatchString(strings.TrimSpace(value.RulesVersion)) {
		return fmt.Errorf("%w: semantic rules version is required", ErrInvalidRecap)
	}
	if !isSHA256Hex(value.RulesDigest) {
		return fmt.Errorf("%w: rules digest is required", ErrInvalidRecap)
	}
	if err := ValidateMetricsForPeriod(value.Metrics, value.Period); err != nil {
		return fmt.Errorf("%w: metrics: %v", ErrInvalidRecap, err)
	}
	if value.Metrics.TotalEvents < minEventsForRecap {
		return fmt.Errorf("%w: total events are below recap minimum", ErrInvalidRecap)
	}
	if err := ValidateStoredRates(value.Metrics); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecap, err)
	}
	if err := ValidateActionableState(value.ActionableState); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRecap, err)
	}
	if err := ValidateBehavior(value.Behavior); err != nil {
		return fmt.Errorf("%w: behavior: %v", ErrInvalidRecap, err)
	}
	if err := ValidateAchievements(value.Achievements); err != nil {
		return fmt.Errorf("%w: achievements: %v", ErrInvalidRecap, err)
	}
	if err := ValidateNextAction(value.NextAction); err != nil {
		return fmt.Errorf("%w: next action: %v", ErrInvalidRecap, err)
	}
	if err := ValidateCards(value.Cards); err != nil {
		return fmt.Errorf("%w: cards: %v", ErrInvalidRecap, err)
	}
	if err := ValidateShareCardConsistency(value); err != nil {
		return fmt.Errorf("%w: share card: %v", ErrInvalidRecap, err)
	}
	if value.GeneratedAt.IsZero() {
		return fmt.Errorf("%w: generated time is required", ErrInvalidRecap)
	}
	if value.GeneratedAt.Before(value.Period.EndAt) {
		return fmt.Errorf("%w: final recap was generated before period completion", ErrInvalidRecap)
	}
	if !value.ActionableState.CapturedAt.Equal(value.GeneratedAt) {
		return fmt.Errorf("%w: actionable-state capture time must equal generated time", ErrInvalidRecap)
	}
	return nil
}

func ValidatePeriod(period model.RecapPeriod) error {
	if period.Year == 0 || period.StartAt.IsZero() || period.EndAt.IsZero() || !period.Final {
		return errors.New("completed annual period is required")
	}
	expectedStart := time.Date(int(period.Year), time.January, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(int(period.Year)+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	if !period.StartAt.Equal(expectedStart) || !period.EndAt.Equal(expectedEnd) {
		return errors.New("period must cover exactly one UTC calendar year")
	}
	return nil
}

func ValidateBehavior(value model.Behavior) error {
	if !model.IsKnownBehaviorCode(value.Code) {
		return fmt.Errorf("unknown code %q", value.Code)
	}
	if value.Title == "" || value.Description == "" || value.Reason == "" {
		return errors.New("text is incomplete")
	}
	if value.Code == model.BehaviorUniversal {
		if value.Score != 0 {
			return errors.New("universal behavior must have score 0")
		}
		return nil
	}
	if value.Score == 0 || len(value.Evidence) == 0 {
		return errors.New("scored behavior requires evidence")
	}
	var score uint64
	for index, item := range value.Evidence {
		if item.Metric == "" || item.Detail == "" || math.IsNaN(item.Actual) || math.IsInf(item.Actual, 0) ||
			math.IsNaN(item.Threshold) || math.IsInf(item.Threshold, 0) {
			return fmt.Errorf("evidence %d is invalid", index)
		}
		score += uint64(item.Points)
	}
	if score != uint64(value.Score) {
		return fmt.Errorf("evidence score %d differs from behavior score %d", score, value.Score)
	}
	return nil
}

func ValidateAchievements(values []model.Achievement) error {
	if len(values) > ruleset.MaxAchievements {
		return fmt.Errorf("too many achievements: got %d, maximum is %d", len(values), ruleset.MaxAchievements)
	}
	seenCodes := make(map[model.AchievementCode]struct{}, len(values))
	for index, value := range values {
		if !model.IsKnownAchievementCode(value.Code) {
			return fmt.Errorf("achievement %d has unknown code %q", index, value.Code)
		}
		if !model.IsKnownAchievementCategory(value.Category) {
			return fmt.Errorf("achievement %d has unknown category %q", index, value.Category)
		}
		if value.Title == "" || value.Description == "" || value.Reason == "" {
			return fmt.Errorf("achievement %d text is incomplete", index)
		}
		if _, ok := seenCodes[value.Code]; ok {
			return fmt.Errorf("duplicate achievement code %q", value.Code)
		}
		seenCodes[value.Code] = struct{}{}
	}
	return nil
}

func ValidateShareCardConsistency(value model.Recap) error {
	if len(value.Cards) == 0 {
		return errors.New("cards are required")
	}
	last := value.Cards[len(value.Cards)-1]
	payload, ok := last.Payload.(model.ShareCard)
	if last.Type != model.CardShare || !ok {
		return errors.New("final card must contain a share-card payload")
	}
	if payload.ShareID != value.ShareID || payload.Year != value.Year {
		return errors.New("share payload is not bound to recap share id and year")
	}
	return nil
}

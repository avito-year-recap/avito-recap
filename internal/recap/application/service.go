package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/analytics"
	"github.com/year-recap/internal/recap/engine"
	"github.com/year-recap/internal/recap/insight"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/narrative"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/validation/structural"
)

type IDGenerator func() (uuid.UUID, error)

type Service struct {
	profiles        ProfileStorage
	analytics       AnalyticsStorage
	actionStates    ActionStateStorage
	recaps          RecapStorage
	events          EventRangeStorage
	engine          *engine.Engine
	narrative       narrative.Enricher
	insight         insight.Generator
	generateFlights generationFlightGroup

	now   func() time.Time
	newID IDGenerator
}

type serviceConfig struct {
	rules     ruleset.Ruleset
	narrative narrative.Enricher
	events    EventRangeStorage
	insight   insight.Generator
	now       func() time.Time
	newID     IDGenerator
}

type Option func(*serviceConfig)

func NewService(
	profiles ProfileStorage,
	analytics AnalyticsStorage,
	actionStates ActionStateStorage,
	recaps RecapStorage,
	options ...Option,
) (*Service, error) {
	if profiles == nil {
		return nil, fmt.Errorf("%w: profile storage", ErrMissingDependency)
	}
	if analytics == nil {
		return nil, fmt.Errorf("%w: analytics storage", ErrMissingDependency)
	}
	if actionStates == nil {
		return nil, fmt.Errorf("%w: action-state storage", ErrMissingDependency)
	}
	if recaps == nil {
		return nil, fmt.Errorf("%w: recap storage", ErrMissingDependency)
	}

	config := serviceConfig{rules: ruleset.DefaultRuleset(), now: time.Now, newID: generateID}
	for _, option := range options {
		if option != nil {
			option(&config)
		}
	}
	core, err := engine.New(config.rules)
	if err != nil {
		return nil, err
	}
	return &Service{
		profiles: profiles, analytics: analytics, actionStates: actionStates, recaps: recaps,
		events: config.events, engine: core, narrative: config.narrative, insight: config.insight,
		now: config.now, newID: config.newID,
	}, nil
}

func (s *Service) ListProfiles(ctx context.Context) ([]model.Profile, error) {
	profiles, err := s.profiles.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	for index, profile := range profiles {
		profile = model.NormalizeProfile(profile)
		if err := structural.ValidateProfile(profile); err != nil {
			return nil, fmt.Errorf("validate profile at index %d: %w", index, err)
		}
		profiles[index] = profile
	}
	return profiles, nil
}

func (s *Service) GetProfileByCode(ctx context.Context, code string) (model.Profile, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return model.Profile{}, ErrProfileNotFound
	}
	profile, err := s.profiles.GetProfileByCode(ctx, code)
	if err != nil {
		return model.Profile{}, fmt.Errorf("get profile by code: %w", err)
	}
	profile = model.NormalizeProfile(profile)
	if profile.Code != code {
		return model.Profile{}, fmt.Errorf("%w: requested %q, got %q", ErrProfileIDMismatch, code, profile.Code)
	}
	return profile, nil
}

var (
	ErrInvalidProfileID  = errors.New("invalid profile id")
	ErrProfileNotFound   = errors.New("profile not found")
	ErrMetricsNotFound   = errors.New("annual metrics not found")
	ErrInvalidRecapID    = errors.New("invalid recap id")
	ErrInvalidShareID    = errors.New("invalid share id")
	ErrNotEnoughActivity = engine.ErrNotEnoughActivity
	ErrMissingDependency = errors.New("missing service dependency")
	ErrGenerateID        = errors.New("generate recap id")
	ErrProfileIDMismatch = errors.New("profile storage returned another profile")
	ErrRecapIDMismatch   = errors.New("recap storage returned another recap")
	ErrShareIDMismatch   = errors.New("recap storage returned another share id")
	ErrRecapKeyMismatch  = errors.New("recap storage returned another idempotency key")
	ErrRecapNotFound     = errors.New("recap not found")

	ErrInvalidYear            = analytics.ErrInvalidYear
	ErrYearNotComplete        = analytics.ErrYearNotComplete
	ErrInvalidProfile         = structural.ErrInvalidProfile
	ErrInvalidMetrics         = structural.ErrInvalidMetrics
	ErrInvalidActionableState = structural.ErrInvalidActionableState
	ErrInvalidRecap           = structural.ErrInvalidRecap

	ErrInvalidPeriod         = errors.New("invalid analysis period")
	ErrPeriodTooLong         = errors.New("analysis period exceeds the maximum supported range")
	ErrEventRangeUnsupported = errors.New("this storage backend cannot analyze an arbitrary date range")
	ErrInsightUnavailable    = errors.New("behavior insight generator is not configured")
	ErrNoActivityInPeriod    = errors.New("no events found for profile in the requested period")
)

func WithClock(clock func() time.Time) Option {
	return func(config *serviceConfig) {
		if clock != nil {
			config.now = clock
		}
	}
}

func WithIDGenerator(generator IDGenerator) Option {
	return func(config *serviceConfig) {
		if generator != nil {
			config.newID = generator
		}
	}
}

func WithRuleset(configured ruleset.Ruleset) Option {
	return func(config *serviceConfig) { config.rules = configured }
}

func WithNarrativeEnricher(enricher narrative.Enricher) Option {
	return func(config *serviceConfig) { config.narrative = enricher }
}

func WithEventRangeStorage(storage EventRangeStorage) Option {
	return func(config *serviceConfig) { config.events = storage }
}

func WithInsightGenerator(generator insight.Generator) Option {
	return func(config *serviceConfig) { config.insight = generator }
}

func (s *Service) generateNonNilID(kind string) (uuid.UUID, error) {
	value, err := s.newID()
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s: %w", ErrGenerateID, kind, err)
	}
	if value == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s: generated nil UUID", ErrGenerateID, kind)
	}
	return value, nil
}

func generateID() (uuid.UUID, error) { return uuid.NewRandom() }

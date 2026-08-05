package recap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidProfileID  = errors.New("invalid profile id")
	ErrInvalidRecapID    = errors.New("invalid recap id")
	ErrInvalidShareID    = errors.New("invalid share id")
	ErrInvalidYear       = errors.New("invalid recap year")
	ErrYearNotComplete   = errors.New("recap year is not complete")
	ErrNotEnoughActivity = errors.New("not enough activity to generate recap")
	ErrMissingDependency = errors.New("missing service dependency")
	ErrGenerateID        = errors.New("generate recap id")
	ErrProfileIDMismatch = errors.New("profile storage returned another profile")
	ErrRecapIDMismatch   = errors.New("recap storage returned another recap")
	ErrShareIDMismatch   = errors.New("recap storage returned another share id")
	ErrRecapKeyMismatch  = errors.New("recap storage returned another idempotency key")
	ErrRecapNotFound     = errors.New("recap not found")
)

const minEventsForRecap uint64 = 5

type IDGenerator func() (uuid.UUID, error)

type Service struct {
	profiles     ProfileStorage
	analytics    AnalyticsStorage
	actionStates ActionStateStorage
	recaps       RecapStorage
	ruleset      Ruleset

	now   func() time.Time
	newID IDGenerator
}

type Option func(*Service)

func WithClock(clock func() time.Time) Option {
	return func(service *Service) {
		if clock != nil {
			service.now = clock
		}
	}
}

func WithIDGenerator(generator IDGenerator) Option {
	return func(service *Service) {
		if generator != nil {
			service.newID = generator
		}
	}
}

func WithRuleset(ruleset Ruleset) Option {
	return func(service *Service) { service.ruleset = ruleset }
}

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

	service := &Service{
		profiles: profiles, analytics: analytics, actionStates: actionStates, recaps: recaps,
		ruleset: DefaultRuleset(), now: time.Now, newID: generateID,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	service.ruleset.Version = normalizeString(service.ruleset.Version)
	service.ruleset.Algorithm = normalizeString(service.ruleset.Algorithm)
	service.ruleset.SharePolicy.Version = normalizeString(service.ruleset.SharePolicy.Version)
	if err := service.ruleset.Validate(); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) ListProfiles(ctx context.Context) ([]Profile, error) {
	profiles, err := s.profiles.ListProfiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	for index, profile := range profiles {
		profile = normalizeProfile(profile)
		if err := validateProfile(profile); err != nil {
			return nil, fmt.Errorf("validate profile at index %d: %w", index, err)
		}
		profiles[index] = profile
	}
	return profiles, nil
}

func (s *Service) Generate(ctx context.Context, profileID uuid.UUID, year uint32) (Recap, error) {
	if profileID == uuid.Nil {
		return Recap{}, ErrInvalidProfileID
	}
	now := s.now().UTC()
	period, err := completedYearPeriod(year, now)
	if err != nil {
		return Recap{}, err
	}
	key := RecapKey{ProfileID: profileID, Year: year, RulesVersion: s.ruleset.Version, RulesDigest: s.ruleset.Digest()}

	if existing, err := s.recaps.GetRecapByKey(ctx, key); err == nil {
		return s.validateStoredByKey(existing, key)
	} else if !errors.Is(err, ErrRecapNotFound) {
		return Recap{}, fmt.Errorf("get recap by idempotency key: %w", err)
	}

	profile, err := s.profiles.GetProfile(ctx, profileID)
	if err != nil {
		return Recap{}, fmt.Errorf("get profile: %w", err)
	}
	profile = normalizeProfile(profile)
	if err := validateProfile(profile); err != nil {
		return Recap{}, err
	}
	if profile.ID != profileID {
		return Recap{}, fmt.Errorf("%w: requested %s, got %s", ErrProfileIDMismatch, profileID, profile.ID)
	}

	metrics, err := s.analytics.CalculateMetrics(ctx, profileID, period)
	if err != nil {
		return Recap{}, fmt.Errorf("calculate metrics: %w", err)
	}
	metrics = normalizeMetrics(metrics)
	if err := validateMetricsForPeriod(metrics, period); err != nil {
		return Recap{}, err
	}
	if metrics.TotalEvents < minEventsForRecap {
		return Recap{}, ErrNotEnoughActivity
	}
	metrics = EnrichMetrics(metrics)

	state, err := s.actionStates.GetActionableState(ctx, profileID, now)
	if err != nil {
		return Recap{}, fmt.Errorf("get actionable state: %w", err)
	}
	state = normalizeActionableState(state)
	if err := validateActionableState(state); err != nil {
		return Recap{}, err
	}
	if !state.CapturedAt.Equal(now) {
		return Recap{}, fmt.Errorf("%w: snapshot captured at %s, requested %s", ErrInvalidActionableState, state.CapturedAt, now)
	}

	behavior := s.ruleset.DetectBehavior(metrics)
	achievements := s.ruleset.BuildAchievements(metrics)
	nextAction := s.ruleset.BuildNextAction(metrics, state, behavior)

	recapID, err := s.generateNonNilID("internal recap")
	if err != nil {
		return Recap{}, err
	}
	shareID, err := s.generateNonNilID("public share")
	if err != nil {
		return Recap{}, err
	}
	if recapID == shareID {
		return Recap{}, fmt.Errorf("%w: internal and public ids must differ", ErrGenerateID)
	}
	cards := BuildCardsWithRuleset(s.ruleset, profile, year, shareID, metrics, behavior, achievements, nextAction)

	candidate := Recap{
		ID: recapID, ShareID: shareID, Profile: profile, Year: year, Period: period,
		RulesVersion: s.ruleset.Version, RulesDigest: s.ruleset.Digest(), Metrics: metrics, ActionableState: state,
		Behavior: behavior, Achievements: achievements, Cards: cards, NextAction: nextAction,
		GeneratedAt: now,
	}
	if err := validateRecapAgainstRuleset(candidate, s.ruleset, now); err != nil {
		return Recap{}, fmt.Errorf("validate generated recap: %w", err)
	}

	// The storage operation is the concurrency boundary. It must insert candidate or
	// atomically return the already stored value for the same unique key.
	stored, err := s.recaps.CreateRecapIfAbsent(ctx, key, candidate)
	if err != nil {
		return Recap{}, fmt.Errorf("create recap if absent: %w", err)
	}
	return s.validateStoredByKey(stored, key)
}

func (s *Service) Get(ctx context.Context, recapID uuid.UUID) (Recap, error) {
	if recapID == uuid.Nil {
		return Recap{}, ErrInvalidRecapID
	}
	value, err := s.recaps.GetRecap(ctx, recapID)
	if err != nil {
		return Recap{}, fmt.Errorf("get recap: %w", err)
	}
	if value.ID != recapID {
		return Recap{}, fmt.Errorf("%w: requested %s, got %s", ErrRecapIDMismatch, recapID, value.ID)
	}
	value = normalizeRecap(value)
	if err := validateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
}

func (s *Service) GetShareCard(ctx context.Context, shareID uuid.UUID) (ShareCard, error) {
	if shareID == uuid.Nil {
		return ShareCard{}, ErrInvalidShareID
	}
	value, err := s.recaps.GetRecapByShareID(ctx, shareID)
	if err != nil {
		return ShareCard{}, fmt.Errorf("get recap by share id: %w", err)
	}
	if value.ShareID != shareID {
		return ShareCard{}, fmt.Errorf("%w: requested %s, got %s", ErrShareIDMismatch, shareID, value.ShareID)
	}
	value = normalizeRecap(value)
	if err := validateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return ShareCard{}, fmt.Errorf("validate shared recap: %w", err)
	}
	return BuildShareCardWithRuleset(s.ruleset, value), nil
}

func (s *Service) validateStoredByKey(value Recap, key RecapKey) (Recap, error) {
	value = normalizeRecap(value)
	if value.Key() != key {
		return Recap{}, fmt.Errorf("%w: requested %+v, got %+v", ErrRecapKeyMismatch, key, value.Key())
	}
	if err := validateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
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

func generateID() (uuid.UUID, error)                     { return uuid.NewRandom() }
func generateIDFrom(reader io.Reader) (uuid.UUID, error) { return uuid.NewRandomFromReader(reader) }

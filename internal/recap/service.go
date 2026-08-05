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
	ErrInvalidYear       = errors.New("invalid recap year")
	ErrNotEnoughActivity = errors.New("not enough activity to generate recap")
	ErrMissingDependency = errors.New("missing service dependency")
	ErrGenerateID        = errors.New("generate recap id")
	ErrProfileIDMismatch = errors.New("profile storage returned another profile")
	ErrRecapIDMismatch   = errors.New("recap storage returned another recap")
)

const minEventsForRecap uint64 = 5

type IDGenerator func() (uuid.UUID, error)

type Service struct {
	profiles  ProfileStorage
	analytics AnalyticsStorage
	recaps    RecapStorage

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

func NewService(
	profiles ProfileStorage,
	analytics AnalyticsStorage,
	recaps RecapStorage,
	options ...Option,
) (*Service, error) {
	if profiles == nil {
		return nil, fmt.Errorf("%w: profile storage", ErrMissingDependency)
	}
	if analytics == nil {
		return nil, fmt.Errorf("%w: analytics storage", ErrMissingDependency)
	}
	if recaps == nil {
		return nil, fmt.Errorf("%w: recap storage", ErrMissingDependency)
	}

	service := &Service{
		profiles:  profiles,
		analytics: analytics,
		recaps:    recaps,
		now:       time.Now,
		newID:     generateID,
	}

	for _, option := range options {
		if option != nil {
			option(service)
		}
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

func (s *Service) Generate(
	ctx context.Context,
	profileID uuid.UUID,
	year uint32,
) (Recap, error) {
	if profileID == uuid.Nil {
		return Recap{}, ErrInvalidProfileID
	}

	now := s.now().UTC()
	if year == 0 || year > uint32(now.Year()) {
		return Recap{}, ErrInvalidYear
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

	metrics, err := s.analytics.CalculateMetrics(ctx, profileID, year)
	if err != nil {
		return Recap{}, fmt.Errorf("calculate metrics: %w", err)
	}
	metrics = normalizeMetrics(metrics)
	if err := validateMetrics(metrics); err != nil {
		return Recap{}, err
	}
	if metrics.TotalEvents < minEventsForRecap {
		return Recap{}, ErrNotEnoughActivity
	}
	metrics = EnrichMetrics(metrics)

	behavior := DetectBehavior(metrics)
	achievements := BuildAchievements(metrics)
	nextAction := BuildNextAction(metrics)
	cards := BuildCards(profile, year, metrics, behavior, achievements, nextAction)

	recapID, err := s.newID()
	if err != nil {
		return Recap{}, fmt.Errorf("%w: %w", ErrGenerateID, err)
	}
	if recapID == uuid.Nil {
		return Recap{}, fmt.Errorf("%w: generated nil UUID", ErrGenerateID)
	}

	value := Recap{
		ID:           recapID,
		Profile:      profile,
		Year:         year,
		RulesVersion: CurrentRulesVersion,
		Metrics:      metrics,
		Behavior:     behavior,
		Achievements: achievements,
		Cards:        cards,
		NextAction:   nextAction,
		GeneratedAt:  now,
	}

	if err := validateRecap(value); err != nil {
		return Recap{}, fmt.Errorf("validate generated recap: %w", err)
	}

	if err := s.recaps.SaveRecap(ctx, value); err != nil {
		return Recap{}, fmt.Errorf("save recap: %w", err)
	}

	return value, nil
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
	if err := validateRecap(value); err != nil {
		return Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}

	return value, nil
}

func (s *Service) GetShareCard(ctx context.Context, recapID uuid.UUID) (ShareCard, error) {
	value, err := s.Get(ctx, recapID)
	if err != nil {
		return ShareCard{}, err
	}

	return BuildShareCard(value), nil
}

func generateID() (uuid.UUID, error) {
	return uuid.NewRandom()
}

func generateIDFrom(reader io.Reader) (uuid.UUID, error) {
	return uuid.NewRandomFromReader(reader)
}

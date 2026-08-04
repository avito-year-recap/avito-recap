package recap

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	ErrInvalidProfileID  = errors.New("profile id is required")
	ErrInvalidRecapID    = errors.New("recap id is required")
	ErrInvalidYear       = errors.New("invalid recap year")
	ErrNotEnoughActivity = errors.New("not enough activity to generate recap")
	ErrMissingDependency = errors.New("missing service dependency")
	ErrGenerateID        = errors.New("generate recap id")
)

const minEventsForRecap uint64 = 5

type IDGenerator func() (string, error)

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
		if err := validateProfile(profile); err != nil {
			return nil, fmt.Errorf("validate profile at index %d: %w", index, err)
		}
	}

	return profiles, nil
}

func (s *Service) Generate(
	ctx context.Context,
	profileID string,
	year uint32,
) (Recap, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
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
	if err := validateProfile(profile); err != nil {
		return Recap{}, err
	}

	metrics, err := s.analytics.CalculateMetrics(ctx, profileID, year)
	if err != nil {
		return Recap{}, fmt.Errorf("calculate metrics: %w", err)
	}
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
	if !isUUID(recapID) {
		return Recap{}, fmt.Errorf("%w: generated id is not a UUID", ErrGenerateID)
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

	if err := s.recaps.SaveRecap(ctx, value); err != nil {
		return Recap{}, fmt.Errorf("save recap: %w", err)
	}

	return value, nil
}

func (s *Service) Get(ctx context.Context, recapID string) (Recap, error) {
	recapID = strings.TrimSpace(recapID)
	if !isUUID(recapID) {
		return Recap{}, ErrInvalidRecapID
	}

	value, err := s.recaps.GetRecap(ctx, recapID)
	if err != nil {
		return Recap{}, fmt.Errorf("get recap: %w", err)
	}

	return value, nil
}

func (s *Service) GetShareCard(ctx context.Context, recapID string) (ShareCard, error) {
	value, err := s.Get(ctx, recapID)
	if err != nil {
		return ShareCard{}, err
	}

	return BuildShareCard(value), nil
}

func generateID() (string, error) {
	return generateIDFrom(rand.Reader)
}

func generateIDFrom(reader io.Reader) (string, error) {
	var raw [16]byte
	if _, err := io.ReadFull(reader, raw[:]); err != nil {
		return "", err
	}

	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	encoded := hex.EncodeToString(raw[:])
	return encoded[0:8] + "-" +
		encoded[8:12] + "-" +
		encoded[12:16] + "-" +
		encoded[16:20] + "-" +
		encoded[20:32], nil
}

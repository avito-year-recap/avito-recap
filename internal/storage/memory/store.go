package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/validation/structural"
	"github.com/year-recap/internal/seed"
)

var ErrInvalidSeedData = errors.New("invalid seed data")

type scenarioKey struct {
	profileID uuid.UUID
	year      uint32
}

// Store is a concurrency-safe demo adapter backed by the repository seed
// catalogue. Generated recaps live only for the lifetime of the process.
type Store struct {
	mu sync.RWMutex

	profiles       []model.Profile
	profilesByID   map[uuid.UUID]model.Profile
	profilesByCode map[string]model.Profile
	metrics        map[scenarioKey]model.Metrics
	actionStates   map[uuid.UUID]model.ActionableState
	recapsByKey    map[model.RecapKey]model.Recap
	recapsByID     map[uuid.UUID]model.Recap
	recapsByShare  map[uuid.UUID]model.Recap
}

func Load(profilesPath, scenariosPath string) (*Store, error) {
	var profiles []model.Profile
	if err := decodeJSONFile(profilesPath, &profiles); err != nil {
		return nil, fmt.Errorf("load profiles: %w", err)
	}
	var scenarios []seed.Scenario
	if err := decodeJSONFile(scenariosPath, &scenarios); err != nil {
		return nil, fmt.Errorf("load scenarios: %w", err)
	}
	return New(profiles, scenarios)
}

func New(profiles []model.Profile, scenarios []seed.Scenario) (*Store, error) {
	if len(profiles) == 0 {
		return nil, fmt.Errorf("%w: profiles are required", ErrInvalidSeedData)
	}
	store := &Store{
		profiles:       make([]model.Profile, 0, len(profiles)),
		profilesByID:   make(map[uuid.UUID]model.Profile, len(profiles)),
		profilesByCode: make(map[string]model.Profile, len(profiles)),
		metrics:        make(map[scenarioKey]model.Metrics, len(scenarios)),
		actionStates:   make(map[uuid.UUID]model.ActionableState, len(scenarios)),
		recapsByKey:    make(map[model.RecapKey]model.Recap),
		recapsByID:     make(map[uuid.UUID]model.Recap),
		recapsByShare:  make(map[uuid.UUID]model.Recap),
	}
	profileIDsByCode := make(map[string]uuid.UUID, len(profiles))
	for index, profile := range profiles {
		profile = model.NormalizeProfile(profile)
		if err := structural.ValidateProfile(profile); err != nil {
			return nil, fmt.Errorf("%w: profile %d: %w", ErrInvalidSeedData, index, err)
		}
		if _, exists := store.profilesByID[profile.ID]; exists {
			return nil, fmt.Errorf("%w: duplicate profile id %s", ErrInvalidSeedData, profile.ID)
		}
		if _, exists := profileIDsByCode[profile.Code]; exists {
			return nil, fmt.Errorf("%w: duplicate profile code %q", ErrInvalidSeedData, profile.Code)
		}
		store.profiles = append(store.profiles, profile)
		store.profilesByID[profile.ID] = profile
		store.profilesByCode[profile.Code] = profile
		profileIDsByCode[profile.Code] = profile.ID
	}
	if len(scenarios) == 0 {
		return nil, fmt.Errorf("%w: scenarios are required", ErrInvalidSeedData)
	}
	for index, scenario := range scenarios {
		profileID, exists := profileIDsByCode[scenario.ProfileCode]
		if !exists {
			return nil, fmt.Errorf("%w: scenario %d references unknown profile %q", ErrInvalidSeedData, index, scenario.ProfileCode)
		}
		key := scenarioKey{profileID: profileID, year: scenario.Year}
		if _, exists := store.metrics[key]; exists {
			return nil, fmt.Errorf("%w: duplicate scenario for profile %q and year %d", ErrInvalidSeedData, scenario.ProfileCode, scenario.Year)
		}
		if _, exists := store.actionStates[profileID]; exists {
			return nil, fmt.Errorf("%w: multiple actionable states for profile %q", ErrInvalidSeedData, scenario.ProfileCode)
		}
		metrics, err := seed.MetricsFromScenario(scenario)
		if err != nil {
			return nil, fmt.Errorf("%w: scenario %d: %w", ErrInvalidSeedData, index, err)
		}
		state := model.NormalizeActionableState(scenario.ActionableState)
		state.CapturedAt = time.Unix(1, 0).UTC()
		if err := structural.ValidateActionableState(state); err != nil {
			return nil, fmt.Errorf("%w: scenario %d: %w", ErrInvalidSeedData, index, err)
		}
		state.CapturedAt = time.Time{}
		store.metrics[key] = metrics
		store.actionStates[profileID] = state
	}
	if len(store.profiles) != len(store.actionStates) {
		return nil, fmt.Errorf(
			"%w: profile/scenario count mismatch: %d/%d",
			ErrInvalidSeedData,
			len(store.profiles),
			len(store.actionStates),
		)
	}
	return store, nil
}

func (s *Store) ListProfiles(ctx context.Context) ([]model.Profile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]model.Profile(nil), s.profiles...), nil
}

func (s *Store) GetProfile(ctx context.Context, profileID uuid.UUID) (model.Profile, error) {
	if err := ctx.Err(); err != nil {
		return model.Profile{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, exists := s.profilesByID[profileID]
	if !exists {
		return model.Profile{}, application.ErrProfileNotFound
	}
	return profile, nil
}

func (s *Store) GetProfileByCode(ctx context.Context, code string) (model.Profile, error) {
	if err := ctx.Err(); err != nil {
		return model.Profile{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, exists := s.profilesByCode[code]
	if !exists {
		return model.Profile{}, application.ErrProfileNotFound
	}
	return profile, nil
}

func (s *Store) CalculateMetrics(
	ctx context.Context,
	profileID uuid.UUID,
	period model.RecapPeriod,
) (model.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return model.Metrics{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	metrics, exists := s.metrics[scenarioKey{profileID: profileID, year: period.Year}]
	if !exists {
		return model.Metrics{}, application.ErrMetricsNotFound
	}
	metrics.CategoryActivities = append([]model.CategoryActivity(nil), metrics.CategoryActivities...)
	return metrics, nil
}

func (s *Store) GetActionableState(
	ctx context.Context,
	profileID uuid.UUID,
	asOf time.Time,
) (model.ActionableState, error) {
	if err := ctx.Err(); err != nil {
		return model.ActionableState{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, exists := s.actionStates[profileID]
	if !exists {
		return model.ActionableState{}, application.ErrRecapNotFound
	}
	state.CapturedAt = asOf.UTC()
	return state, nil
}

func (s *Store) GetRecapByKey(ctx context.Context, key model.RecapKey) (model.Recap, error) {
	if err := ctx.Err(); err != nil {
		return model.Recap{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.recapsByKey[key]
	if !exists {
		return model.Recap{}, application.ErrRecapNotFound
	}
	return value, nil
}

func (s *Store) CreateRecapIfAbsent(
	ctx context.Context,
	key model.RecapKey,
	value model.Recap,
) (model.Recap, error) {
	if err := ctx.Err(); err != nil {
		return model.Recap{}, err
	}
	if value.Key() != key {
		return model.Recap{}, fmt.Errorf("%w: recap key mismatch", ErrInvalidSeedData)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, exists := s.recapsByKey[key]; exists {
		return existing, nil
	}
	if _, exists := s.recapsByID[value.ID]; exists {
		return model.Recap{}, fmt.Errorf("internal recap id collision: %s", value.ID)
	}
	if _, exists := s.recapsByShare[value.ShareID]; exists {
		return model.Recap{}, fmt.Errorf("public share id collision: %s", value.ShareID)
	}
	s.recapsByKey[key] = value
	s.recapsByID[value.ID] = value
	s.recapsByShare[value.ShareID] = value
	return value, nil
}

func (s *Store) GetRecap(ctx context.Context, recapID uuid.UUID) (model.Recap, error) {
	if err := ctx.Err(); err != nil {
		return model.Recap{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.recapsByID[recapID]
	if !exists {
		return model.Recap{}, application.ErrRecapNotFound
	}
	return value, nil
}

func (s *Store) GetRecapByShareID(ctx context.Context, shareID uuid.UUID) (model.Recap, error) {
	if err := ctx.Err(); err != nil {
		return model.Recap{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.recapsByShare[shareID]
	if !exists {
		return model.Recap{}, application.ErrRecapNotFound
	}
	return value, nil
}

func decodeJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

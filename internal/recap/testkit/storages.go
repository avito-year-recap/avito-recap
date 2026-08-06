package testkit

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/application"
	"github.com/year-recap/internal/recap/model"
)

type ProfileStorage struct {
	Profile model.Profile
}

func (s *ProfileStorage) ListProfiles(context.Context) ([]model.Profile, error) {
	return []model.Profile{s.Profile}, nil
}
func (s *ProfileStorage) GetProfile(context.Context, uuid.UUID) (model.Profile, error) {
	return s.Profile, nil
}

type AnalyticsStorage struct {
	Metrics model.Metrics
	Calls   int
}

func (s *AnalyticsStorage) CalculateMetrics(context.Context, uuid.UUID, model.RecapPeriod) (model.Metrics, error) {
	s.Calls++
	return s.Metrics, nil
}

type ActionStateStorage struct{ State model.ActionableState }

func (s *ActionStateStorage) GetActionableState(_ context.Context, _ uuid.UUID, asOf time.Time) (model.ActionableState, error) {
	value := s.State
	if value.CapturedAt.IsZero() {
		value.CapturedAt = asOf
	}
	return value, nil
}

type RecapStorage struct {
	mu      sync.Mutex
	Value   *model.Recap
	Creates int
}

func (s *RecapStorage) GetRecapByKey(_ context.Context, key model.RecapKey) (model.Recap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Value != nil && s.Value.Key() == key {
		return *s.Value, nil
	}
	return model.Recap{}, application.ErrRecapNotFound
}
func (s *RecapStorage) CreateRecapIfAbsent(_ context.Context, _ model.RecapKey, value model.Recap) (model.Recap, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Value != nil {
		return *s.Value, nil
	}
	s.Creates++
	copyValue := value
	s.Value = &copyValue
	return copyValue, nil
}
func (s *RecapStorage) GetRecap(_ context.Context, id uuid.UUID) (model.Recap, error) {
	if s.Value != nil && s.Value.ID == id {
		return *s.Value, nil
	}
	return model.Recap{}, application.ErrRecapNotFound
}
func (s *RecapStorage) GetRecapByShareID(_ context.Context, id uuid.UUID) (model.Recap, error) {
	if s.Value != nil && s.Value.ShareID == id {
		return *s.Value, nil
	}
	return model.Recap{}, application.ErrRecapNotFound
}

package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/ruleset"
	"github.com/year-recap/internal/recap/validation/structural"
	"time"
)

type IDGenerator func() (uuid.UUID, error)

type Service struct {
	profiles     ProfileStorage
	analytics    AnalyticsStorage
	actionStates ActionStateStorage
	recaps       RecapStorage
	ruleset      ruleset.Ruleset

	now   func() time.Time
	newID IDGenerator
}

type Option func(*Service)

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
		ruleset: ruleset.DefaultRuleset(), now: time.Now, newID: generateID,
	}
	for _, option := range options {
		if option != nil {
			option(service)
		}
	}
	service.ruleset.Version = model.NormalizeString(service.ruleset.Version)
	service.ruleset.Algorithm = model.NormalizeString(service.ruleset.Algorithm)
	service.ruleset.SharePolicy.Version = model.NormalizeString(service.ruleset.SharePolicy.Version)
	if err := service.ruleset.Validate(); err != nil {
		return nil, err
	}
	return service, nil
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

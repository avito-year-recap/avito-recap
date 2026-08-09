package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

type ProfileStorage interface {
	ListProfiles(ctx context.Context) ([]model.Profile, error)
	GetProfile(ctx context.Context, profileID uuid.UUID) (model.Profile, error)
	GetProfileByCode(ctx context.Context, code string) (model.Profile, error)
}

type AnalyticsStorage interface {
	CalculateMetrics(ctx context.Context, profileID uuid.UUID, period model.RecapPeriod) (model.Metrics, error)
}

// ActionStateStorage returns an addressable point-in-time snapshot captured at asOf.
type ActionStateStorage interface {
	GetActionableState(ctx context.Context, profileID uuid.UUID, asOf time.Time) (model.ActionableState, error)
}

// RecapStorage owns the idempotency and uniqueness boundary for generated recaps.
type RecapStorage interface {
	GetRecapByKey(ctx context.Context, key model.RecapKey) (model.Recap, error)
	CreateRecapIfAbsent(ctx context.Context, key model.RecapKey, value model.Recap) (model.Recap, error)
	GetRecap(ctx context.Context, recapID uuid.UUID) (model.Recap, error)
	GetRecapByShareID(ctx context.Context, shareID uuid.UUID) (model.Recap, error)
}

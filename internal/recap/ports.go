package recap

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProfileStorage interface {
	ListProfiles(ctx context.Context) ([]Profile, error)
	GetProfile(ctx context.Context, profileID uuid.UUID) (Profile, error)
}

type AnalyticsStorage interface {
	CalculateMetrics(ctx context.Context, profileID uuid.UUID, period RecapPeriod) (Metrics, error)
}

// ActionStateStorage must return an addressable snapshot captured exactly at asOf.
// Positive draft/dialog/listing counts require a representative non-nil ID.
type ActionStateStorage interface {
	GetActionableState(ctx context.Context, profileID uuid.UUID, asOf time.Time) (ActionableState, error)
}

// RecapStorage must enforce a unique constraint on
// (profile_id, year, rules_version, rules_digest), unique internal/share IDs,
// and atomically implement CreateRecapIfAbsent.
type RecapStorage interface {
	GetRecapByKey(ctx context.Context, key RecapKey) (Recap, error)
	CreateRecapIfAbsent(ctx context.Context, key RecapKey, value Recap) (Recap, error)
	GetRecap(ctx context.Context, recapID uuid.UUID) (Recap, error)
	GetRecapByShareID(ctx context.Context, shareID uuid.UUID) (Recap, error)
}

package recap

import (
	"context"

	"github.com/google/uuid"
)

// ProfileStorage hides the concrete database from the domain service.
// The ClickHouse implementation will live in internal/storage/clickhouse.
type ProfileStorage interface {
	ListProfiles(ctx context.Context) ([]Profile, error)
	GetProfile(ctx context.Context, profileID uuid.UUID) (Profile, error)
}

type AnalyticsStorage interface {
	CalculateMetrics(
		ctx context.Context,
		profileID uuid.UUID,
		year uint32,
	) (Metrics, error)
}

type RecapStorage interface {
	SaveRecap(ctx context.Context, value Recap) error
	GetRecap(ctx context.Context, recapID uuid.UUID) (Recap, error)
}

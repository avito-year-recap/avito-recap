package connect

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/insight"
	"github.com/year-recap/internal/recap/model"
)

type Application interface {
	ListProfiles(context.Context) ([]model.Profile, error)
	GetProfileByCode(context.Context, string) (model.Profile, error)
	Generate(context.Context, uuid.UUID, uint32) (model.Recap, error)
	ExplainRecap(context.Context, uuid.UUID, uint32) (model.RecapExplanation, error)
	Get(context.Context, uuid.UUID) (model.Recap, error)
	GetShareCard(context.Context, uuid.UUID) (model.ShareCard, error)
	AnalyzeBehavior(ctx context.Context, profileID uuid.UUID, start, end time.Time) (insight.Result, error)
}

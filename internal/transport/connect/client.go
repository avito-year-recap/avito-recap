package connect

import (
	"context"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

type Application interface {
	ListProfiles(context.Context) ([]model.Profile, error)
	GetProfileByCode(context.Context, string) (model.Profile, error)
	Generate(context.Context, uuid.UUID, uint32) (model.Recap, error)
	Get(context.Context, uuid.UUID) (model.Recap, error)
	GetShareCard(context.Context, uuid.UUID) (model.ShareCard, error)
}

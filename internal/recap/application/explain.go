package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

// ExplainRecap returns a decision trace for the same immutable recap addressed
// by profile/year. Generate is idempotent, so this does not create a different
// result when the recap already exists.
func (s *Service) ExplainRecap(ctx context.Context, profileID uuid.UUID, year uint32) (model.RecapExplanation, error) {
	value, err := s.Generate(ctx, profileID, year)
	if err != nil {
		return model.RecapExplanation{}, err
	}
	return s.engine.Explain(value), nil
}

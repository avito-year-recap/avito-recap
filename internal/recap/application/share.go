package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/integrity"
	"github.com/year-recap/internal/recap/model"
	"github.com/year-recap/internal/recap/presentation/share"
)

func (s *Service) GetShareCard(ctx context.Context, shareID uuid.UUID) (model.ShareCard, error) {
	if shareID == uuid.Nil {
		return model.ShareCard{}, ErrInvalidShareID
	}
	value, err := s.recaps.GetRecapByShareID(ctx, shareID)
	if err != nil {
		return model.ShareCard{}, fmt.Errorf("get recap by share id: %w", err)
	}
	if value.ShareID != shareID {
		return model.ShareCard{}, fmt.Errorf("%w: requested %s, got %s", ErrShareIDMismatch, shareID, value.ShareID)
	}
	value = model.NormalizeRecap(value)
	if err := integrity.ValidateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return model.ShareCard{}, fmt.Errorf("validate shared recap: %w", err)
	}
	return share.BuildWithRuleset(s.ruleset, value), nil
}

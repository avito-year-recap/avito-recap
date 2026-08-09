package application

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/integrity"
	"github.com/year-recap/internal/recap/model"
)

func (s *Service) Get(ctx context.Context, recapID uuid.UUID) (model.Recap, error) {
	if recapID == uuid.Nil {
		return model.Recap{}, ErrInvalidRecapID
	}
	value, err := s.recaps.GetRecap(ctx, recapID)
	if err != nil {
		return model.Recap{}, fmt.Errorf("get recap: %w", err)
	}
	if value.ID != recapID {
		return model.Recap{}, fmt.Errorf("%w: requested %s, got %s", ErrRecapIDMismatch, recapID, value.ID)
	}
	value = model.NormalizeRecap(value)
	if err := integrity.ValidateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return model.Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
}

func (s *Service) validateStoredByKey(value model.Recap, key model.RecapKey) (model.Recap, error) {
	value = model.NormalizeRecap(value)
	if value.Key() != key {
		return model.Recap{}, fmt.Errorf("%w: requested %+v, got %+v", ErrRecapKeyMismatch, key, value.Key())
	}
	if err := integrity.ValidateRecapAgainstRuleset(value, s.ruleset, s.now().UTC()); err != nil {
		return model.Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
}

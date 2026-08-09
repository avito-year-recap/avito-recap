package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	value, err = s.engine.ValidateStored(value, s.now().UTC())
	if err != nil {
		return model.Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
}

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
	value, err = s.engine.ValidateStored(value, s.now().UTC())
	if err != nil {
		return model.ShareCard{}, fmt.Errorf("validate shared recap: %w", err)
	}
	return s.engine.PublicProjection(value), nil
}

func (s *Service) validateStoredByKey(value model.Recap, key model.RecapKey) (model.Recap, error) {
	value = model.NormalizeRecap(value)
	if value.Key() != key {
		return model.Recap{}, fmt.Errorf("%w: requested %+v, got %+v", ErrRecapKeyMismatch, key, value.Key())
	}
	value, err := s.engine.ValidateStored(value, s.now().UTC())
	if err != nil {
		return model.Recap{}, fmt.Errorf("validate stored recap: %w", err)
	}
	return value, nil
}

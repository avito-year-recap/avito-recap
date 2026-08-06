package structural

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
	"strings"
)

func ValidateProfile(profile model.Profile) error {
	if profile.ID == uuid.Nil {
		return fmt.Errorf("%w: id is required", ErrInvalidProfile)
	}
	if strings.TrimSpace(profile.Code) == "" {
		return fmt.Errorf("%w: code is required", ErrInvalidProfile)
	}
	if strings.TrimSpace(profile.DisplayName) == "" {
		return fmt.Errorf("%w: display name is required", ErrInvalidProfile)
	}
	return nil
}

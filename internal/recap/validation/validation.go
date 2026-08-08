package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/year-recap/internal/recap/model"
)

var (
	ErrInvalidProfile         = errors.New("invalid profile")
	ErrInvalidMetrics         = errors.New("invalid metrics")
	ErrInvalidActionableState = errors.New("invalid actionable state")
	ErrInvalidRecap           = errors.New("invalid recap")
)

var (
	categoryCodePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	semanticVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
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

func isSafeCategoryCode(code string) bool {
	return categoryCodePattern.MatchString(strings.TrimSpace(code))
}

func isSafeApplicationRoute(route string) bool {
	switch route {
	case "/favorites", "/listings/new", "/recommendations":
		return true
	default:
		return false
	}
}

func isSHA256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func sumUint64(values ...uint64) (uint64, bool) {
	var result uint64
	for _, value := range values {
		if value > ^uint64(0)-result {
			return 0, false
		}
		result += value
	}
	return result, true
}

func rate(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total)
}

package structural

import (
	"regexp"
	"strings"
)

var (
	categoryCodePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)
	semanticVersionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)
)

const minEventsForRecap uint64 = 5

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
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

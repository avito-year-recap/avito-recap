package share

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isSafePublicText(value string, maxRunes int) bool {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > maxRunes {
		return false
	}
	for _, r := range value {
		if unicode.IsControl(r) || unicode.Is(unicode.Bidi_Control, r) {
			return false
		}
	}
	return true
}

var categoryCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

func isSafeCategoryCode(code string) bool {
	return categoryCodePattern.MatchString(strings.TrimSpace(code))
}

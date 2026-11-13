package validator

import (
	"strings"
	"unicode"
)

// hasControlOrInvisible reports whether s contains Unicode Cc/Cf characters
// (allowing '\n' when allowNewline is true).
func hasControlOrInvisible(s string, allowNewline bool) bool {
	for _, r := range s {
		if allowNewline && r == '\n' {
			continue
		}
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return true
		}
	}

	return false
}

func hasSpoofingChars(s string) bool {
	return strings.ContainsAny(s, "@#:`")
}

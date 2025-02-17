package lexer

import (
	"regexp"
)

const (
	CHAR                string = `[^\n\x60]`
	NON_WHITESPACE_CHAR string = `[^\s\x60]`
	INLINE_WHITESPACE   string = `[^\S\t\r\n]`
)

func patternBuilder(patterns ...string) *regexp.Regexp {
	finalPattern := ``

	for _, pattern := range patterns {
		finalPattern += pattern
	}

	return regexp.MustCompile(finalPattern)
}

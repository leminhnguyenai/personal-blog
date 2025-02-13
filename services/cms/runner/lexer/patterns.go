package lexer

import (
	"regexp"
)

const (
	PARAGRAPH         string = `[^\[\]\(\)\n\x60]`
	CHARACTER         string = `[^\[\]\(\)\s\x60]`
	INDENTABLE        string = `^[^\S\t\r\n]*?`
	INLINE_WHITESPACE string = `[^\S\t\r\n]`
)

func patternBuilder(patterns ...string) *regexp.Regexp {
	finalPattern := ``

	for _, pattern := range patterns {
		finalPattern += pattern
	}

	return regexp.MustCompile(finalPattern)
}

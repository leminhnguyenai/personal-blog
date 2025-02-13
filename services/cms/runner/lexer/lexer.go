package lexer

import (
	"fmt"
	"regexp"
	"strings"
)

// Function to modifier the lexer
type regexHandler func(lex *lexer, regex *regexp.Regexp)

// Contain the regex pattern for identifying and the handler to modify the lexer
type regexPattern struct {
	regex   *regexp.Regexp
	handler regexHandler
}

type lexer struct {
	patterns []regexPattern
	Tokens   []Token
	source   string
	pos      int
}

// Advance the current position of the lexer by n
func (lex *lexer) advanceN(n int) {
	lex.pos += n
}

func (lex *lexer) push(token Token) {
	lex.Tokens = append(lex.Tokens, token)
}

func (lex *lexer) at() byte {
	return lex.source[lex.pos]
}

// Return the rest of the source string
func (lex *lexer) remainder() string {
	return lex.source[lex.pos:]
}

func (lex *lexer) at_eof() bool {
	return lex.pos >= len(lex.source)
}

// Get the xy-plane location of the current position
func (lex *lexer) getLoc(pos int, separator string) []int {
	tokenizedString := lex.source[:pos]
	lines := strings.Split(tokenizedString, separator)
	lastLine := lines[len(lines)-1]

	return []int{len(lines) - 1, len(lastLine)}
}

// Handler for tokens that have a fixed length
// The handler remove the left side whitespace of the matched string
func defaultHandler(kind TokenKind, pattern string) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		matchString := regex.FindString(lex.remainder())
		actualToken := patternBuilder(pattern).FindString(matchString)

		startLoc := lex.getLoc(lex.pos+len(matchString)-len(actualToken), "\n")
		lex.advanceN(len(matchString))
		endLoc := lex.getLoc(lex.pos-1, "\n")

		lex.push(NewToken(kind, NewLoc(startLoc, endLoc), actualToken))
	}
}

// Handler for tokens that doesn't have a predefined length (e.g string)
// This handler will take the whole matched string as the value
func dynamicHandler(kind TokenKind) regexHandler {
	return func(lex *lexer, regex *regexp.Regexp) {
		matchString := regex.FindString(lex.remainder())
		startLoc := lex.getLoc(lex.pos, "\n")
		lex.advanceN(len(matchString))
		endLoc := lex.getLoc(lex.pos-1, "\n")

		lex.push(NewToken(kind, NewLoc(startLoc, endLoc), matchString))
	}
}

// Handler to skip through newlines
func skipHandler(lex *lexer, regex *regexp.Regexp) {
	matchLoc := regex.FindStringIndex(lex.remainder())
	lex.advanceN(matchLoc[1])
}

func linkHandler(lex *lexer, regex *regexp.Regexp) {
	matchString := regex.FindString(lex.remainder())
	placeholder := patternBuilder(`\[`, PARAGRAPH, `*`, `\]`).FindString(matchString)
	link := patternBuilder(`\(`, PARAGRAPH, `*`, `\)`).FindString(matchString)

	placeholder = placeholder[1 : len(placeholder)-1]
	link = link[1 : len(link)-1]

	startLoc := lex.getLoc(lex.pos, "\n")
	lex.advanceN(len(matchString))
	endLoc := lex.getLoc(lex.pos-1, "\n")

	lex.push(NewToken(LINK, NewLoc(startLoc, endLoc), placeholder, link))
}

func paragraphHandler(lex *lexer, regex *regexp.Regexp) {
	prevChar := lex.source[lex.pos-1]

	if string(prevChar) == "\n" {
		dynamicHandler(NEWLINE_PARAGRAPH)(lex, regex)
	} else {
		dynamicHandler(INLINE_PARAGRAPH)(lex, regex)
	}
}

func CreateLexer(source string) *lexer {
	return &lexer{
		source: source,
		patterns: []regexPattern{
			{patternBuilder(`\n`, `+`), skipHandler},
			{patternBuilder(INDENTABLE, `\d+\.\s*`), dynamicHandler(NUMBERED_LIST)},
			{patternBuilder(`\[`, PARAGRAPH, `*`, `\]`, `\(`, PARAGRAPH, `*`, `\)`), linkHandler},
			{patternBuilder(INLINE_WHITESPACE, `*`, `#####\s`), defaultHandler(HEADING_5, `#####\s`)},
			{patternBuilder(INLINE_WHITESPACE, `*`, `####\s`), defaultHandler(HEADING_4, `####\s`)},
			{patternBuilder(INLINE_WHITESPACE, `*`, `###\s`), defaultHandler(HEADING_3, `###\s`)},
			{patternBuilder(INLINE_WHITESPACE, `*`, `##\s`), defaultHandler(HEADING_2, `##\s`)},
			{patternBuilder(INLINE_WHITESPACE, `*`, `#\s`), defaultHandler(HEADING_1, `#\s`)},
			{patternBuilder(INDENTABLE, `-\s`), dynamicHandler(DASH)},
			{patternBuilder(PARAGRAPH, `+`), paragraphHandler},
		},
	}
}

func Tokenize(source string) ([]Token, error) {
	lex := CreateLexer(source)

	for !lex.at_eof() {
		matched := false

		for _, pattern := range lex.patterns { // Iterate the pattern, not the source string
			// Find the first location that match the pattern
			loc := pattern.regex.FindStringIndex(lex.remainder())
			// Only match if the location is found AND the match location is at the beginning of the string (not source string) that is iterated
			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}

		if !matched {
			loc := lex.getLoc(lex.pos, "\n")

			return nil, fmt.Errorf(
				"Lexer::error -> unrecognized token near\n %s\nat [%d:%d]\n",
				lex.remainder(), loc[0], loc[1],
			)
		}
	}

	return lex.Tokens, nil
}

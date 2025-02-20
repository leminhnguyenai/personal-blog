package lexer

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Common patterns
	CHAR                = `[^\n]`
	NON_WHITESPACE_CHAR = `[^\s]`
	INLINE_WHITESPACE   = `[^\S\t\r\n]`
	INDENT              = INLINE_WHITESPACE + `*`

	// Actual patterns for tokens
	SKIP_NEWLINE_PATTERN = `\n+`

	HEADING_5_PATTERN = INDENT + `#####` + INLINE_WHITESPACE
	HEADING_4_PATTERN = INDENT + `####` + INLINE_WHITESPACE
	HEADING_3_PATTERN = INDENT + `###` + INLINE_WHITESPACE
	HEADING_2_PATTERN = INDENT + `##` + INLINE_WHITESPACE
	HEADING_1_PATTERN = INDENT + `#` + INLINE_WHITESPACE

	NUMBERED_LIST_PATTERN = INDENT + `\d+\.` + INLINE_WHITESPACE
	DASH_PATTERN          = INDENT + `-` + INLINE_WHITESPACE

	CODEBLOCK_PATTERN = `\x60\x60\x60` + `(.|\n)*` + `\x60\x60\x60`

	CALLOUT_NOTE_PATTERN      = INDENT + `>\s\[!NOTE\]` + INDENT
	CALLOUT_IMPORTANT_PATTERN = INDENT + `>\s\[!IMPORTANT\]` + INDENT
	CALLOUT_WARNING_PATTERN   = INDENT + `>\s\[!WARNING\]` + INLINE_WHITESPACE
	CALLOUT_EXAMPLE_PATTERN   = INDENT + `>\s\[!EXAMPLE\]` + INLINE_WHITESPACE

	QUOTE_PATTERN = INDENT + `>` + INLINE_WHITESPACE

	PARAGRAPH_PATTERN = CHAR + `+`
)

// Function to modifier the lexer
type patternHandler func(lex *lexer, regex string)

// Contain the regex pattern for identifying and the handler to modify the lexer
type regexPattern struct {
	regex   string
	handler patternHandler
}

type lexer struct {
	patterns []regexPattern
	tokens   []Token
	source   string
	pos      int
}

// Advance the current position of the lexer by n
func (lex *lexer) advanceN(n int) {
	lex.pos += n
}

func (lex *lexer) push(token Token) {
	lex.tokens = append(lex.tokens, token)
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
func (lex *lexer) getLoc(pos int) [2]int {
	tokenizedString := lex.source[:pos]
	lines := strings.Split(tokenizedString, "\n")
	lastLine := lines[len(lines)-1]

	return [2]int{len(lines) - 1, len(lastLine)}
}

func (lex *lexer) isOnNewLine() bool {
	return lex.pos == 0 || string(lex.source[lex.pos-1]) == "\n"
}

// Handler for tokens that have a fixed length
// The handler remove the left side whitespace of the matched string
func blockHandler(kind TokenKind) patternHandler {
	return func(lex *lexer, regex string) {
		if lex.isOnNewLine() ||
			(kind != HEADING_1 &&
				kind != HEADING_2 &&
				kind != HEADING_3 &&
				kind != HEADING_4 &&
				kind != HEADING_5 &&
				lex.tokens[len(lex.tokens)-1].kind == QUOTE) {
			matchString := regexp.MustCompile(regex).FindString(lex.remainder())
			rightside_indent := regexp.MustCompile(`^` + INLINE_WHITESPACE + `*`).FindString(matchString)

			startLoc := lex.getLoc(lex.pos + len(rightside_indent))
			lex.advanceN(len(matchString))
			endLoc := lex.getLoc(lex.pos - 1)

			lex.push(NewToken(kind, NewLoc(startLoc, endLoc), matchString[len(rightside_indent):]))
		} else {
			paragraphHandler(lex, PARAGRAPH_PATTERN)
		}
	}
}

// Handler for tokens that doesn't have a predefined length (e.g string)
// This handler will take the whole matched string as the value
func paragraphHandler(lex *lexer, regex string) {
	matchString := regexp.MustCompile(regex).FindString(lex.remainder())
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchString))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(PARAGRAPH, NewLoc(startLoc, endLoc), matchString))
}

// Handler to skip through newlines
func skipHandler(lex *lexer, regex string) {
	matchLoc := regexp.MustCompile(regex).FindStringIndex(lex.remainder())
	lex.advanceN(matchLoc[1])
}

func codeBlockHandler(lex *lexer, regex string) {
	if lex.pos == 0 || string(lex.source[lex.pos-1]) == "\n" {
		matchString := regexp.MustCompile(regex).FindString(lex.remainder())
		fileType := strings.ToLower(strings.Split(matchString, "\n")[0][3:])
		code := strings.Split(matchString, "\n")
		code = code[1 : len(code)-1]

		startLoc := lex.getLoc(lex.pos)
		lex.advanceN(len(matchString))
		endLoc := lex.getLoc(lex.pos - 1)

		lex.push(NewToken(CODE_BLOCK, NewLoc(startLoc, endLoc), fileType, strings.Join(code, "\n")))

	} else {
		paragraphHandler(lex, PARAGRAPH_PATTERN)
	}
}

// COMMIT: Switch to use paragraph token as value for callout
func CreateLexer(source string) *lexer {
	return &lexer{
		source: source,
		patterns: []regexPattern{
			{SKIP_NEWLINE_PATTERN, skipHandler},
			{(HEADING_5_PATTERN), blockHandler(HEADING_5)},
			{(HEADING_4_PATTERN), blockHandler(HEADING_4)},
			{(HEADING_3_PATTERN), blockHandler(HEADING_3)},
			{(HEADING_2_PATTERN), blockHandler(HEADING_2)},
			{(HEADING_1_PATTERN), blockHandler(HEADING_1)},
			{(NUMBERED_LIST_PATTERN), blockHandler(NUMBERED_LIST)},
			{(DASH_PATTERN), blockHandler(DASH)},
			{(CODEBLOCK_PATTERN), codeBlockHandler},
			{(CALLOUT_NOTE_PATTERN), blockHandler(CALLOUT_NOTE)},
			{(CALLOUT_IMPORTANT_PATTERN), blockHandler(CALLOUT_IMPORTANT)},
			{(CALLOUT_WARNING_PATTERN), blockHandler(CALLOUT_WARNING)},
			{(CALLOUT_EXAMPLE_PATTERN), blockHandler(CALLOUT_EXAMPLE)},
			{(QUOTE_PATTERN), blockHandler(QUOTE)},
			{(PARAGRAPH_PATTERN), paragraphHandler},
		},
	}
}

func Tokenize(source string) ([]Token, error) {
	lex := CreateLexer(source)

	for !lex.at_eof() {
		matched := false

		// Iterate the pattern, not the source string
		for _, pattern := range lex.patterns {
			// Find the first location that match the pattern
			loc := regexp.MustCompile(pattern.regex).FindStringIndex(lex.remainder())
			// Only match if the location is found AND the match location is
			// at the beginning of the string (not source string) that is iterated
			// NOTE: This remove the need for "^" regex pattern
			if loc != nil && loc[0] == 0 {
				pattern.handler(lex, pattern.regex)
				matched = true
				break
			}
		}

		if !matched {
			loc := lex.getLoc(lex.pos)

			return nil, fmt.Errorf(
				"Lexer::error -> unrecognized token near\n %s\nat [%d:%d]\n",
				lex.remainder(), loc[0], loc[1],
			)
		}
	}

	return lex.tokens, nil
}

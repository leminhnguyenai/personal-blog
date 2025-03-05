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

	// Pattern for block elements
	SKIP_NEWLINE_PATTERN = `\n+`

	// NOTE: Due to Go's lack of negative lockahead, frontmatter currently won't support hyphen in the content
	FRONTMATTER_PATTERN = `(^---)([^-]*)(\n---)`

	HEADING_5_PATTERN = INLINE_WHITESPACE + `*` + `#####` + INLINE_WHITESPACE
	HEADING_4_PATTERN = INLINE_WHITESPACE + `*` + `####` + INLINE_WHITESPACE
	HEADING_3_PATTERN = INLINE_WHITESPACE + `*` + `###` + INLINE_WHITESPACE
	HEADING_2_PATTERN = INLINE_WHITESPACE + `*` + `##` + INLINE_WHITESPACE
	HEADING_1_PATTERN = INLINE_WHITESPACE + `*` + `#` + INLINE_WHITESPACE

	NUMBERED_LIST_PATTERN = INLINE_WHITESPACE + `*` + `\d+\.` + INLINE_WHITESPACE
	HYPHEN_LIST_PATTERN   = INLINE_WHITESPACE + `*` + `-` + INLINE_WHITESPACE

	CODEBLOCK_DELIMITER_PATTERN = `\x60\x60\x60` + CHAR + `*`

	CALLOUT_NOTE_PATTERN      = INLINE_WHITESPACE + `*` + `>\s\[!NOTE\]` + INLINE_WHITESPACE + `*`
	CALLOUT_IMPORTANT_PATTERN = INLINE_WHITESPACE + `*` + `>\s\[!IMPORTANT\]` + INLINE_WHITESPACE + `*`
	CALLOUT_WARNING_PATTERN   = INLINE_WHITESPACE + `*` + `>\s\[!WARNING\]` + INLINE_WHITESPACE + `*`
	CALLOUT_EXAMPLE_PATTERN   = INLINE_WHITESPACE + `*` + `>\s\[!EXAMPLE\]` + INLINE_WHITESPACE + `*`

	QUOTE_PATTERN = INLINE_WHITESPACE + `*` + `>` + INLINE_WHITESPACE

	PARAGRAPH_PATTERN = CHAR + `+`

	// Patterns for inline elements
	LINK_PATTERN        = `\[[^\n\[\]\(\)]*\]\([^\n\[\]\(\)]*\)`
	INLINE_CODE_PATTERN = `\x60` + CHAR + `*` + `\x60`
)

type match func(lex *lexer) string

type patternHandler func(lex *lexer, matchStr string)

type patternConstructor struct {
	match   match
	handler patternHandler
}

type lexer struct {
	constructors []patternConstructor
	tokens       []Token
	source       string
	pos          int
}

func NewLexer(source string, constructors []patternConstructor) *lexer {
	return &lexer{
		source:       source,
		constructors: constructors,
	}
}

// Advance the current position of the lexer by n
func (lex *lexer) advanceN(n int) {
	lex.pos += n
}

func (lex *lexer) push(token Token) {
	lex.tokens = append(lex.tokens, token)
}

// Return the rest of the source string
func (lex *lexer) remainder() string {
	return lex.source[lex.pos:]
}

func (lex *lexer) at_eof() bool {
	return lex.pos >= len(lex.source)
}

func (lex *lexer) isOnNewLine() bool {
	return lex.pos == 0 || string(lex.source[lex.pos-1]) == "\n"
}

// Get the xy-plane location of the current position
func (lex *lexer) getLoc(pos int) [2]int {
	tokenizedString := lex.source[:pos]
	lines := strings.Split(tokenizedString, "\n")
	lastLine := lines[len(lines)-1]

	return [2]int{len(lines) - 1, len(lastLine)}
}

func blockTokenMatch(regex string) match {
	return func(lex *lexer) string {
		matchLoc := regexp.MustCompile(regex).FindStringIndex(lex.remainder())

		if matchLoc != nil && matchLoc[0] == 0 && lex.isOnNewLine() {
			return lex.remainder()[matchLoc[0]:matchLoc[1]]
		} else {
			return ""
		}
	}
}

func blockTokenHandler(kind TokenKind) patternHandler {
	return func(lex *lexer, matchStr string) {
		rightside_indent := regexp.MustCompile(`^` + INLINE_WHITESPACE + `*`).FindString(matchStr)

		startLoc := lex.getLoc(lex.pos + len(rightside_indent))
		lex.advanceN(len(matchStr))
		endLoc := lex.getLoc(lex.pos - 1)

		lex.push(NewToken(kind, NewLoc(startLoc, endLoc), matchStr[len(rightside_indent):]))
	}
}

func paragraphMatch(lex *lexer) string {
	matchLoc := regexp.MustCompile(PARAGRAPH_PATTERN).FindStringIndex(lex.remainder())

	if matchLoc != nil && matchLoc[0] == 0 {
		return lex.remainder()[matchLoc[0]:matchLoc[1]]
	} else {
		return ""
	}
}

func paragraphHandler(lex *lexer, matchStr string) {
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(PARAGRAPH, NewLoc(startLoc, endLoc), matchStr))
}

func skipLinesMatch(lex *lexer) string {
	matchLoc := regexp.MustCompile(SKIP_NEWLINE_PATTERN).FindStringIndex(lex.remainder())

	if matchLoc != nil && matchLoc[0] == 0 {
		return lex.remainder()[matchLoc[0]:matchLoc[1]]
	} else {
		return ""
	}
}

func skipLinesHandler(lex *lexer, matchStr string) {
	lex.advanceN(len(matchStr))
}

// COMMIT: Rewrite the lexer with support for block tokens
func Tokenize(source string) ([]Token, error) {
	lex := NewLexer(source, []patternConstructor{
		{skipLinesMatch, skipLinesHandler},
		{blockTokenMatch(HEADING_5_PATTERN), blockTokenHandler(HEADING_5)},
		{blockTokenMatch(HEADING_4_PATTERN), blockTokenHandler(HEADING_4)},
		{blockTokenMatch(HEADING_3_PATTERN), blockTokenHandler(HEADING_3)},
		{blockTokenMatch(HEADING_2_PATTERN), blockTokenHandler(HEADING_2)},
		{blockTokenMatch(HEADING_1_PATTERN), blockTokenHandler(HEADING_1)},
		{blockTokenMatch(HYPHEN_LIST_PATTERN), blockTokenHandler(HYPHEN_LIST)},
		{blockTokenMatch(NUMBERED_LIST_PATTERN), blockTokenHandler(NUMBERED_LIST)},
		{paragraphMatch, paragraphHandler},
	})

	for !lex.at_eof() {
		for _, constructor := range lex.constructors {
			matchStr := constructor.match(lex)

			// NOTE: Match should assure that both the match status and the match location
			if matchStr != "" {
				constructor.handler(lex, matchStr)
				goto CONTINUE
			}
		}
		goto ERROR

	CONTINUE:
		continue
	ERROR:
		loc := lex.getLoc(lex.pos)

		return nil, fmt.Errorf(
			"Lexer::error -> unrecognized token near\n %s\nat [%d:%d]\n",
			lex.remainder(), loc[0], loc[1],
		)
	}

	return lex.tokens, nil
}

package lexer

import (
	"fmt"
	"regexp"
	"strings"
)

type regexPattern string

func (r regexPattern) isMatchAtFirst(str string) bool {
	loc := regexp.MustCompile(string(r)).FindStringIndex(str)
	return loc != nil && loc[0] == 0
}

func (r regexPattern) findString(str string) string {
	return regexp.MustCompile(string(r)).FindString(str)
}

func (r regexPattern) findStringIndex(str string) []int {
	return regexp.MustCompile(string(r)).FindStringIndex(str)
}

const (
	// Common patterns
	CHAR                regexPattern = `[^\n]`
	NON_WHITESPACE_CHAR regexPattern = `[^\s]`
	INLINE_WHITESPACE   regexPattern = `[^\S\t\r\n]`
	INDENT                           = INLINE_WHITESPACE + `*`

	// Pattern for block elements
	SKIP_NEWLINE_PATTERN regexPattern = `\n+`

	HEADING_5_PATTERN = INDENT + `#####` + INLINE_WHITESPACE
	HEADING_4_PATTERN = INDENT + `####` + INLINE_WHITESPACE
	HEADING_3_PATTERN = INDENT + `###` + INLINE_WHITESPACE
	HEADING_2_PATTERN = INDENT + `##` + INLINE_WHITESPACE
	HEADING_1_PATTERN = INDENT + `#` + INLINE_WHITESPACE

	NUMBERED_LIST_PATTERN = INDENT + `\d+\.` + INLINE_WHITESPACE
	HYPHEN_LIST_PATTERN   = INDENT + `-` + INLINE_WHITESPACE

	CODEBLOCK_DELIMITER_PATTERN regexPattern = `\x60\x60\x60` + CHAR + `*`

	CALLOUT_NOTE_PATTERN      = INDENT + `>\s\[!NOTE\]` + INDENT
	CALLOUT_IMPORTANT_PATTERN = INDENT + `>\s\[!IMPORTANT\]` + INDENT
	CALLOUT_WARNING_PATTERN   = INDENT + `>\s\[!WARNING\]` + INLINE_WHITESPACE
	CALLOUT_EXAMPLE_PATTERN   = INDENT + `>\s\[!EXAMPLE\]` + INLINE_WHITESPACE

	QUOTE_PATTERN = INDENT + `>` + INLINE_WHITESPACE

	PARAGRAPH_PATTERN = CHAR + `+`

	// Patterns for inline elements
	LINK_PATTERN        regexPattern = `\[[^\n\[\]\(\)]*\]\([^\n\[\]\(\)]*\)`
	INLINE_CODE_PATTERN regexPattern = `\x60` + CHAR + `*` + `\x60`
)

// Function to modifier the lexer
type regexHandler func(lex *lexer, pattern regexPattern)

// Contain the regex pattern for identifying and the handler to modify the lexer
type regexConstructor struct {
	regex   regexPattern
	handler regexHandler
}

// Range over the regex constructor and only yield once when the pattern is match at first
func matchPatternAtFirst(lex *lexer, arr []regexConstructor) func(func(regexHandler, regexPattern) bool) {
	return func(yield func(regexHandler, regexPattern) bool) {
		for _, constructor := range arr {
			if constructor.regex.isMatchAtFirst(lex.remainder()) {
				yield(constructor.handler, constructor.regex)
				return
			}
		}
	}
}

type lexer struct {
	patterns []regexConstructor
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

func (lex *lexer) isChildOfQuote(kind TokenKind) bool {
	return (kind != HEADING_1 &&
		kind != HEADING_2 &&
		kind != HEADING_3 &&
		kind != HEADING_4 &&
		kind != HEADING_5 &&
		lex.tokens[len(lex.tokens)-1].kind == QUOTE)
}

// Handler for tokens that have a fixed length
// The handler remove the left side whitespace of the matched string
func blockHandler(kind TokenKind) regexHandler {
	return func(lex *lexer, pattern regexPattern) {
		if lex.isOnNewLine() || lex.isChildOfQuote(kind) {
			matchString := pattern.findString(lex.remainder())
			rightside_indent := regexPattern(`^` + INLINE_WHITESPACE + `*`).findString(matchString)

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
func paragraphHandler(lex *lexer, pattern regexPattern) {
	matchString := pattern.findString(lex.remainder())
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchString))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(PARAGRAPH, NewLoc(startLoc, endLoc), matchString))
}

// Handler to skip through newlines
func skipHandler(lex *lexer, pattern regexPattern) {
	lex.advanceN(pattern.findStringIndex(lex.remainder())[1])
}

// FIX: Codeblock can't detect the boundary
func codeBlockDelimiterHandler(lex *lexer, pattern regexPattern) {
	if lex.isOnNewLine() || lex.isChildOfQuote(CODE_BLOCK) {
		matchString := pattern.findString(lex.remainder())
		fileType := strings.ToLower(strings.Split(matchString, "\n")[0][3:])
		// code := strings.Split(matchString, "\n")
		// code = code[1 : len(code)-1]

		startLoc := lex.getLoc(lex.pos)
		lex.advanceN(len(matchString))
		endLoc := lex.getLoc(lex.pos - 1)

		lex.push(NewToken(CODE_BLOCK, NewLoc(startLoc, endLoc), fileType))

	} else {
		paragraphHandler(lex, PARAGRAPH_PATTERN)
	}
}

func linkHandler(lex *lexer, pattern regexPattern) {
	matchString := pattern.findString(lex.remainder())
	placeholder := regexPattern(`\[` + CHAR + `*` + `\]`).findString(matchString)
	link := regexPattern(`\(` + CHAR + `*` + `\)`).findString(matchString)

	placeholder = placeholder[1 : len(placeholder)-1]
	link = link[1 : len(link)-1]

	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchString))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(LINK, NewLoc(startLoc, endLoc), placeholder, link))
}

func inlineCodeHandler(lex *lexer, pattern regexPattern) {
	matchString := pattern.findString(lex.remainder())

	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchString))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(INLINE_CODE, NewLoc(startLoc, endLoc), matchString[1:len(matchString)-1]))
}

func NewLexer(source string) *lexer {
	return &lexer{
		source: source,
		patterns: []regexConstructor{
			{SKIP_NEWLINE_PATTERN, skipHandler},
			{HEADING_5_PATTERN, blockHandler(HEADING_5)},
			{HEADING_4_PATTERN, blockHandler(HEADING_4)},
			{HEADING_3_PATTERN, blockHandler(HEADING_3)},
			{HEADING_2_PATTERN, blockHandler(HEADING_2)},
			{HEADING_1_PATTERN, blockHandler(HEADING_1)},
			{NUMBERED_LIST_PATTERN, blockHandler(NUMBERED_LIST)},
			{HYPHEN_LIST_PATTERN, blockHandler(HYPHEN_LIST)},
			{CODEBLOCK_DELIMITER_PATTERN, codeBlockDelimiterHandler},
			{CALLOUT_NOTE_PATTERN, blockHandler(CALLOUT_NOTE)},
			{CALLOUT_IMPORTANT_PATTERN, blockHandler(CALLOUT_IMPORTANT)},
			{CALLOUT_WARNING_PATTERN, blockHandler(CALLOUT_WARNING)},
			{CALLOUT_EXAMPLE_PATTERN, blockHandler(CALLOUT_EXAMPLE)},
			{QUOTE_PATTERN, blockHandler(QUOTE)},
			{PARAGRAPH_PATTERN, paragraphHandler},
		},
	}
}

func codeBlockMerger(tokens []Token, startIndex int) (Token, int, int) {
	codeBlock := tokens[startIndex]
	codeBlock.values = append(codeBlock.values, "")

	for i := startIndex + 1; i < len(tokens); i++ {
		if tokens[i].kind == CODE_BLOCK {
			codeBlock.loc.end = tokens[i].loc.end
			return codeBlock, startIndex, i
		}

		codeBlock.values[len(codeBlock.values)-1] += tokens[i].values[0] + "\n"
	}

	return codeBlock, startIndex, len(tokens) - 1
}

// Scan and merge tokens that supposed to be one together
func blockTokensScanner(tokens []Token) []Token {
	cpy := []Token{}

	for i := 0; i < len(tokens); i++ {
		// In this code, a merged token will return the start and end index of the range where
		// tokens are merged, which will be used to appropriately plug into the copy slice,
		// the reason for this is that some merged tokens (like table) not only merge the below tokens
		// but also the above tokens
		switch tokens[i].kind {
		case CODE_BLOCK:
			mergedToken, start, end := codeBlockMerger(tokens, i)
			if len(cpy) >= start {
				cpy = append(cpy[:start], mergedToken)
			} else {
				cpy = append(cpy, mergedToken)
			}
			i = end
		default:
			cpy = append(cpy, tokens[i])
		}
	}

	return cpy
}

// Scan and split paragraph toksn into multiple inline tokens
func inlineTokensScanner(source string, paraLoc [2]int) []Token {
	lex := &lexer{
		source: source,
		patterns: []regexConstructor{
			{LINK_PATTERN, linkHandler},
			{INLINE_CODE_PATTERN, inlineCodeHandler},
		},
	}
	prevLoc := [2]int{0, 0}

	for !lex.at_eof() {
		match := false

		for handler, regex := range matchPatternAtFirst(lex, lex.patterns) {
			currentLoc := lex.getLoc(lex.pos)

			if currentLoc[1] != prevLoc[1] {
				lex.push(NewToken(
					PARAGRAPH,
					NewLoc(prevLoc, [2]int{0, currentLoc[1] - 1}),
					source[prevLoc[1]:currentLoc[1]],
				))
			}

			handler(lex, regex)
			prevLoc = lex.getLoc(lex.pos)
			match = true
		}

		if !match {
			lex.pos++
		}
	}

	if prevLoc[1] < len(source) {
		lex.push(NewToken(
			PARAGRAPH,
			NewLoc(prevLoc, [2]int{0, len(source) - 1}),
			source[prevLoc[1]:],
		))
	}

	for i := range lex.tokens {
		lex.tokens[i].loc.start[0] = paraLoc[0]
		lex.tokens[i].loc.end[0] = paraLoc[0]
		lex.tokens[i].loc.start[1] += paraLoc[1]
		lex.tokens[i].loc.end[1] += paraLoc[1]
	}

	return lex.tokens
}

func Tokenize(source string) ([]Token, error) {
	lex := NewLexer(source)

	for !lex.at_eof() {
		match := false

		for handler, regex := range matchPatternAtFirst(lex, lex.patterns) {
			handler(lex, regex)
			match = true
		}

		if !match {
			loc := lex.getLoc(lex.pos)

			return nil, fmt.Errorf(
				"Lexer::error -> unrecognized token near\n %s\nat [%d:%d]\n",
				lex.remainder(), loc[0], loc[1],
			)
		}
	}

	newTokens := []Token{}

	for _, token := range blockTokensScanner(lex.tokens) {
		fmt.Println(token.kind)
		if token.kind == PARAGRAPH {
			newTokens = append(newTokens, inlineTokensScanner(token.values[0], token.loc.start)...)
		} else {
			newTokens = append(newTokens, token)
		}
	}

	return newTokens, nil
}

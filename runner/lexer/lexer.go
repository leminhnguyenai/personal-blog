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
	INLINE_CODE_PATTERN = `\x60` + `[^\n\x60]` + `*` + `\x60`
)

type patternMatch func(lex *lexer) string

type patternHandler func(lex *lexer, matchStr string)

type patternConstructor struct {
	match   patternMatch
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

func (lex *lexer) isChildOfQuote() bool {
	return lex.tokens[len(lex.tokens)-1].Kind == QUOTE
}

// Get the xy-plane location of the current position
func (lex *lexer) getLoc(pos int) [2]int {
	tokenizedString := lex.source[:pos]
	lines := strings.Split(tokenizedString, "\n")
	lastLine := lines[len(lines)-1]

	return [2]int{len(lines) - 1, len(lastLine)}
}

func blockTokenMatch(regex string) patternMatch {
	return func(lex *lexer) string {
		matchLoc := regexp.MustCompile(regex).FindStringIndex(lex.remainder())

		if matchLoc != nil && matchLoc[0] == 0 && (lex.isOnNewLine() || lex.isChildOfQuote()) {
			return lex.remainder()[matchLoc[0]:matchLoc[1]]
		} else {
			return ""
		}
	}
}

func inlineTokenMatch(regex string) patternMatch {
	return func(lex *lexer) string {
		matchLoc := regexp.MustCompile(regex).FindStringIndex(lex.remainder())

		if matchLoc != nil && matchLoc[0] == 0 {
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

func headingMatch(regex string) patternMatch {
	return func(lex *lexer) string {
		matchLoc := regexp.MustCompile(regex).FindStringIndex(lex.remainder())

		if matchLoc != nil && matchLoc[0] == 0 && lex.isOnNewLine() {
			return lex.remainder()[matchLoc[0]:matchLoc[1]]
		} else {
			return ""
		}
	}
}

func inlineCodeHandler(lex *lexer, matchStr string) {
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(INLINE_CODE, NewLoc(startLoc, endLoc), matchStr[1:len(matchStr)-1]))
}

func boldTextHandler(lex *lexer, matchStr string) {
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(BOLD_TEXT, NewLoc(startLoc, endLoc), matchStr[2:len(matchStr)-2]))
}

func italicTextHandler(lex *lexer, matchStr string) {
	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(ITALIC_TEXT, NewLoc(startLoc, endLoc), matchStr[1:len(matchStr)-1]))
}

func linkHandler(lex *lexer, matchStr string) {
	placeholder := regexp.MustCompile(`\[` + CHAR + `*` + `\]`).FindString(matchStr)
	link := regexp.MustCompile(`\(` + CHAR + `*` + `\)`).FindString(matchStr)

	placeholder = placeholder[1 : len(placeholder)-1]
	link = link[1 : len(link)-1]

	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(LINK, NewLoc(startLoc, endLoc), placeholder, link))
}

func leftsideWhitespacesHandler(lex *lexer, matchStr string) {
	lex.advanceN(len(matchStr))
}

func paragraphHandler(lex *lexer, matchStr string) {
	rightside_indent := regexp.MustCompile(`^` + INLINE_WHITESPACE + `*`).FindString(matchStr)
	startLoc := lex.getLoc(lex.pos + len(rightside_indent))

	if lex.isOnNewLine() || lex.isChildOfQuote() {
		lex.push(NewToken(PARAGRAPH, NewLoc(startLoc, startLoc), matchStr[len(rightside_indent):]))
	}

	lex.advanceN(len(matchStr))

	inlineLex := NewLexer(matchStr[len(rightside_indent):], []patternConstructor{
		{inlineTokenMatch(`\s*$`), leftsideWhitespacesHandler},
		{inlineTokenMatch(INLINE_CODE_PATTERN), inlineCodeHandler},
		{inlineTokenMatch(LINK_PATTERN), linkHandler},
		{inlineTokenMatch(`\*\*[^\n\*]*\*\*`), boldTextHandler},
		{inlineTokenMatch(`_[^\n_]*_`), italicTextHandler},
	})

	prevLoc := 0

	for !inlineLex.at_eof() {
		for _, constructor := range inlineLex.constructors {
			inlineMatchStr := constructor.match(inlineLex)
			if inlineMatchStr != "" {
				currentLoc := inlineLex.pos

				if currentLoc != prevLoc {
					inlineLex.push(
						NewToken(TEXT, NewLoc([2]int{0, prevLoc}, [2]int{0, currentLoc - 1}),
							inlineLex.source[prevLoc:currentLoc]))
				}

				constructor.handler(inlineLex, inlineMatchStr)
				prevLoc = inlineLex.pos
				goto CONTINUE
			}
		}
		inlineLex.pos++
		continue

	CONTINUE:
		continue
	}

	if prevLoc < len(inlineLex.source) {
		inlineLex.push(NewToken(
			TEXT,
			NewLoc([2]int{0, prevLoc}, [2]int{0, len(inlineLex.source) - 1}),
			inlineLex.source[prevLoc:],
		))
	}

	for i := range inlineLex.tokens {
		inlineLex.tokens[i].Loc.start[0] = startLoc[0]
		inlineLex.tokens[i].Loc.end[0] = startLoc[0]
		inlineLex.tokens[i].Loc.start[1] += startLoc[1]
		inlineLex.tokens[i].Loc.end[1] += startLoc[1]
	}

	lex.tokens = append(lex.tokens, inlineLex.tokens...)
}

func codeBlockMatch(lex *lexer) string {
	codeBlockDelimPattern := `\x60\x60\x60` + INLINE_WHITESPACE + `*` + `[a-zA-Z\.-_]+`
	codeBlockDelimBeginLoc := regexp.MustCompile(codeBlockDelimPattern).FindStringIndex(lex.remainder())

	if codeBlockDelimBeginLoc != nil && codeBlockDelimBeginLoc[0] == 0 &&
		(lex.isOnNewLine() || lex.isChildOfQuote()) {
		matchStr := lex.remainder()[codeBlockDelimBeginLoc[0]:codeBlockDelimBeginLoc[1]] + "\n"

		lines := strings.Split(lex.remainder(), "\n")[1:]
		for _, line := range lines {
			if regexp.MustCompile(`^\x60\x60\x60`).FindString(line) != "" {
				matchStr += line
				return matchStr
			}

			matchStr += line + "\n"
		}

		return matchStr
	} else {
		return ""
	}
}

func codeBlockHandler(lex *lexer, matchStr string) {
	lines := strings.Split(matchStr, "\n")
	metadata := regexp.MustCompile(`[a-zA-Z\.-_]+`).FindString(lines[0])
	code := strings.Join(lines[1:len(lines)-1], "\n")

	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(CODE_BLOCK, NewLoc(startLoc, endLoc), metadata, code))
}

func frontmatterMatch(lex *lexer) string {
	horizontalRuleLoc := regexp.MustCompile(`---`).FindStringIndex(lex.remainder())

	if horizontalRuleLoc != nil && horizontalRuleLoc[0] == 0 && lex.isOnNewLine() {
		matchStr := lex.remainder()[horizontalRuleLoc[0]:horizontalRuleLoc[1]] + "\n"

		lines := strings.Split(lex.remainder(), "\n")[1:]
		for _, line := range lines {
			if regexp.MustCompile(`^---`).FindString(line) != "" {
				matchStr += line
				return matchStr
			}

			matchStr += line + "\n"
		}

		return matchStr
	} else {
		return ""
	}
}

func frontmatterHandler(lex *lexer, matchStr string) {
	lines := strings.Split(matchStr, "\n")
	linesOfContent := lines[1 : len(lines)-1]

	values := []string{}

	for i, line := range linesOfContent {
		propertyName := regexp.MustCompile(`^[a-zA-Z]+:`).FindString(line)

		if propertyName == "" {
			continue
		}

		switch propertyName[:len(propertyName)-1] {
		case "id":
			values = append(values, "id", line[4:])
		case "date":
			values = append(values, "date", line[6:])
		case "tags":
			tags := ""

			for j := i + 1; j < len(linesOfContent); j++ {
				str := regexp.MustCompile(INLINE_WHITESPACE + `+` + `- [a-zA-Z0-9]+`).
					FindString(linesOfContent[j])
				tagName := regexp.MustCompile(`[a-zA-Z0-9]+`).FindString(str)

				if tagName == "" {
					break
				}

				tags += tagName + ","
			}

			values = append(values, "tags", tags)
		}
	}

	startLoc := lex.getLoc(lex.pos)
	lex.advanceN(len(matchStr))
	endLoc := lex.getLoc(lex.pos - 1)

	lex.push(NewToken(FRONTMATTER, NewLoc(startLoc, endLoc), values...))
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

func Tokenize(source string) ([]Token, error) {
	lex := NewLexer(source, []patternConstructor{
		{frontmatterMatch, frontmatterHandler},
		{skipLinesMatch, skipLinesHandler},
		{headingMatch(HEADING_5_PATTERN), blockTokenHandler(HEADING_5)},
		{headingMatch(HEADING_4_PATTERN), blockTokenHandler(HEADING_4)},
		{headingMatch(HEADING_3_PATTERN), blockTokenHandler(HEADING_3)},
		{headingMatch(HEADING_2_PATTERN), blockTokenHandler(HEADING_2)},
		{headingMatch(HEADING_1_PATTERN), blockTokenHandler(HEADING_1)},
		{blockTokenMatch(HYPHEN_LIST_PATTERN), blockTokenHandler(HYPHEN_LIST)},
		{blockTokenMatch(NUMBERED_LIST_PATTERN), blockTokenHandler(NUMBERED_LIST)},
		{codeBlockMatch, codeBlockHandler},
		{blockTokenMatch(CALLOUT_NOTE_PATTERN), blockTokenHandler(CALLOUT_NOTE)},
		{blockTokenMatch(CALLOUT_IMPORTANT_PATTERN), blockTokenHandler(CALLOUT_IMPORTANT)},
		{blockTokenMatch(CALLOUT_WARNING_PATTERN), blockTokenHandler(CALLOUT_WARNING)},
		{blockTokenMatch(CALLOUT_EXAMPLE_PATTERN), blockTokenHandler(CALLOUT_EXAMPLE)},
		{blockTokenMatch(QUOTE_PATTERN), blockTokenHandler(QUOTE)},
		{inlineTokenMatch(PARAGRAPH_PATTERN), paragraphHandler},
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

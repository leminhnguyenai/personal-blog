package markdownparser

import (
	"fmt"
	"strings"
)

type TokenKind int

const (
	FRONTMATTER TokenKind = iota

	HEADING_1
	HEADING_2
	HEADING_3
	HEADING_4
	HEADING_5

	CALLOUT_NOTE
	CALLOUT_IMPORTANT
	CALLOUT_WARNING
	CALLOUT_EXAMPLE

	QUOTE

	HYPHEN_LIST
	NUMBERED_LIST
	CODE_BLOCK

	PARAGRAPH

	TEXT
	LINK
	INLINE_CODE
	BOLD_TEXT
	ITALIC_TEXT
)

func getString(vals []string) string {
	cpy := make([]string, len(vals))
	copy(cpy, vals)

	for i := range cpy {
		cpy[i] = strings.Replace(cpy[i], "\n", "\\n", -1)
	}

	if len(cpy) == 1 {
		return cpy[0]
	}

	str := ""

	for i, val := range cpy {
		str += val
		if i < len(cpy)-1 {
			str += " - "
		}
	}

	return str
}

type Location struct {
	start [2]int
	end   [2]int
}

func NewLoc(start, end [2]int) Location {
	return Location{
		start: start,
		end:   end,
	}
}

func (loc Location) Display() string {
	return fmt.Sprintf("    [%d,%d] - [%d,%d]", loc.start[0], loc.start[1], loc.end[0], loc.end[1])
}

type Token struct {
	Kind   TokenKind
	Values []string
	Loc    Location
}

func NewToken(kind TokenKind, loc Location, values ...string) Token {
	return Token{
		Kind:   kind,
		Values: values,
		Loc:    loc,
	}
}

func (token Token) isOneOfKinds(kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if token.Kind == kind {
			return true
		}
	}

	return false
}

// Calculate the length of indentation
func (token Token) Indentation() int {
	return token.Loc.start[1]
}

func (token Token) Debug() string {
	locDisplay := fmt.Sprintf("%s", token.Loc.Display())

	if token.isOneOfKinds(
		NUMBERED_LIST,
		LINK,
		INLINE_CODE,
		BOLD_TEXT,
		ITALIC_TEXT,
		CODE_BLOCK,
		FRONTMATTER,
		TEXT,
	) {
		return fmt.Sprintf("%s (%s)", TokenKindString(token.Kind), getString(token.Values)) + locDisplay
	} else if token.isOneOfKinds(FRONTMATTER) {
		return fmt.Sprintf("%s ()", TokenKindString(token.Kind))
	} else {
		return fmt.Sprintf("%s ()", TokenKindString(token.Kind)) + locDisplay
	}
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case FRONTMATTER:
		return "frontmatter"
	case CALLOUT_NOTE:
		return "callout_note"
	case CALLOUT_IMPORTANT:
		return "callout_important"
	case CALLOUT_WARNING:
		return "callout_warning"
	case CALLOUT_EXAMPLE:
		return "callout_example"
	case QUOTE:
		return "quote"
	case PARAGRAPH:
		return "paragraph"
	case TEXT:
		return "text"
	case LINK:
		return "link"
	case INLINE_CODE:
		return "inline_code"
	case BOLD_TEXT:
		return "bold text"
	case ITALIC_TEXT:
		return "italic text"
	case HEADING_1:
		return "heading_1"
	case HEADING_2:
		return "heading_2"
	case HEADING_3:
		return "heading_3"
	case HEADING_4:
		return "heading_4"
	case HEADING_5:
		return "heading_5"
	case HYPHEN_LIST:
		return "hyphen_list"
	case NUMBERED_LIST:
		return "numbered_list"
	case CODE_BLOCK:
		return "code_block"
	default:
		return ""
	}
}

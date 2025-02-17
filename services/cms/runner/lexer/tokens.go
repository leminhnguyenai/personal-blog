package lexer

import (
	"fmt"
)

type TokenKind int

const (
	SOURCE_FILE TokenKind = iota

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

	DASH
	NUMBERED_LIST
	CODE_BLOCK

	PARAGRAPH

	LINK
	INLINE_CODE
)

type values []string

func (vals values) getString() string {
	if len(vals) == 1 {
		return vals[0]
	}

	str := ""

	for i, val := range vals {
		str += val
		if i < len(vals)-1 {
			str += " - "
		}
	}

	return str
}

type Location struct {
	start []int
	end   []int
}

func NewLoc(start, end []int) Location {
	return Location{
		start: start,
		end:   end,
	}
}

func (loc Location) Display() string {
	return fmt.Sprintf("    [%d,%d] - [%d,%d]", loc.start[0], loc.start[1], loc.end[0], loc.end[1])
}

type Token struct {
	kind   TokenKind
	values values
	loc    Location
}

func NewToken(kind TokenKind, loc Location, values ...string) Token {
	return Token{
		kind:   kind,
		values: values,
		loc:    loc,
	}
}

func (token Token) isOneOfKinds(kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if token.kind == kind {
			return true
		}
	}

	return false
}

// Calculate the length of indentation
func (token Token) Indentation() int {
	return token.loc.start[1]
}

func (token Token) Debug() string {
	locDisplay := fmt.Sprintf("%s", token.loc.Display())

	if token.isOneOfKinds(
		PARAGRAPH,
		NUMBERED_LIST,
		LINK,
		INLINE_CODE,
		CALLOUT_NOTE,
		CALLOUT_IMPORTANT,
		CALLOUT_WARNING,
		CALLOUT_EXAMPLE,
		CODE_BLOCK,
	) {
		return fmt.Sprintf("%s (%s)", TokenKindString(token.kind), token.values.getString()) + locDisplay
	} else if token.isOneOfKinds(SOURCE_FILE) {
		return fmt.Sprintf("%s ()", TokenKindString(token.kind))
	} else {
		return fmt.Sprintf("%s ()", TokenKindString(token.kind)) + locDisplay
	}
}

func TokenKindString(kind TokenKind) string {
	switch kind {
	case SOURCE_FILE:
		return "source_file"
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
	case LINK:
		return "link"
	case INLINE_CODE:
		return "inline_code"
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
	case DASH:
		return "dash"
	case NUMBERED_LIST:
		return "numbered_list"
	case CODE_BLOCK:
		return "code_block"
	default:
		return ""
	}
}

package runner

import (
	"fmt"

	"github.com/leminhnguyenai/personal-blog/services/cms/runner/lexer"
)

func Traverse(node *lexer.Node) (string, string) {
	values := ""
	children := ""

	for _, value := range node.Values {
		switch value.Self.Kind {
		case lexer.PARAGRAPH:
			values += inlineParagraphRenderer(value)
		case lexer.LINK:
			values += linkRenderer(value)
		case lexer.INLINE_CODE:
			values += inlineCodeRenderer(value)
		}
	}

	for _, child := range node.Children {
		switch child.Self.Kind {
		case lexer.HEADING_1, lexer.HEADING_2, lexer.HEADING_3, lexer.HEADING_4, lexer.HEADING_5:
			children += headingRenderer(child)
		case lexer.PARAGRAPH:
			children += paragraphRenderer(child)
		case lexer.HYPHEN_LIST, lexer.NUMBERED_LIST:
			children += listRenderer(child)
		}
	}

	return values, children
}

func headingRenderer(node *lexer.Node) string {
	values, children := Traverse(node)

	return fmt.Sprintf("<h%d>%s</h%d>%s", node.Self.Kind, values, node.Self.Kind, children)
}

func paragraphRenderer(node *lexer.Node) string {
	return fmt.Sprintf(`<p>%s</p>`, node.Self.Values[0])
}

func listRenderer(node *lexer.Node) string {
	values, children := Traverse(node)

	if children != "" {
		children = "<ul>" + children + "</ul>"
	}

	var listNotation string

	if node.Self.Kind == lexer.HYPHEN_LIST {
		listNotation = `<span class="list">‑  </span>`
	} else if node.Self.Kind == lexer.NUMBERED_LIST {
		listNotation = `<span class="list">` + node.Self.Values[0] + `.  </span>`
	}

	return fmt.Sprintf(`<li><span class="list">%s</span>%s%s</li>`, listNotation, values, children)
}

func inlineParagraphRenderer(node *lexer.Node) string {
	return fmt.Sprintf(`<span>%s</span>`, node.Self.Values[0])
}

func linkRenderer(node *lexer.Node) string {
	return fmt.Sprintf(`<a href="%s">%s</a>`, node.Self.Values[1], node.Self.Values[0])
}

func inlineCodeRenderer(node *lexer.Node) string {
	return fmt.Sprintf(`<code>%s</code>`, node.Self.Values[0])
}

package runner

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/leminhnguyenai/personal-blog/services/cms/runner/lexer"
)

func Traverse(node *lexer.Node) (string, string) {
	values := ""
	children := ""

	if node.Self.Kind == lexer.FRONTMATTER {
		values += frontmatterRenderer(node)
	}

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
		case lexer.CODE_BLOCK:
			children += codeBlockRenderer(child)
		case lexer.QUOTE:
			children += quoteRenderer(child)
		case lexer.CALLOUT_NOTE, lexer.CALLOUT_IMPORTANT, lexer.CALLOUT_WARNING, lexer.CALLOUT_EXAMPLE:
			children += calloutRenderer(child)
		}
	}

	return values, children
}

func frontmatterRenderer(node *lexer.Node) string {
	data := ""

	for i := 0; i < len(node.Self.Values); i += 2 {
		propertyName := node.Self.Values[i]
		propertyValue := node.Self.Values[i+1]
		switch propertyName {
		case "id", "date":
			data += fmt.Sprintf(
				"<p><span>%s</span><span>: </span><span>%s</span></p>",
				propertyName,
				propertyValue,
			)
		case "tags":
			tags := strings.Split(propertyValue[1:len(propertyValue)-1], ",")
			for i := range tags {
				tags[i] = strings.Replace(tags[i], `"`, ``, -1)
			}

			data += fmt.Sprintf(
				"<p><span>%s</span><span>: </span><span>%s</span></p>",
				propertyName,
				strings.Join(tags, "  "),
			)
		}

	}

	return "<hr>" + data + "<hr>"
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

func quoteRenderer(node *lexer.Node) string {
	_, children := Traverse(node)

	return fmt.Sprintf(`<blockquote>%s</blockquote>`, children)
}

func calloutRenderer(node *lexer.Node) string {
	values, children := Traverse(node)

	if values == "" {
		switch node.Self.Kind {
		case lexer.CALLOUT_NOTE:
			values = "Note"
		case lexer.CALLOUT_IMPORTANT:
			values = "Important"
		case lexer.CALLOUT_WARNING:
			values = "Warning"
		case lexer.CALLOUT_EXAMPLE:
			values = "Example"
		}
	}

	return fmt.Sprintf(`<div><p>%s</p>%s</div>`, values, children)
}

func codeBlockRenderer(node *lexer.Node) string {
	return fmt.Sprintf(
		`<pre data-lang="%s">%s</pre>`,
		node.Self.Values[0],
		template.HTMLEscapeString(node.Self.Values[1]),
	)
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

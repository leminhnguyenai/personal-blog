package renderer

import (
	"html/template"
	"strings"

	"github.com/leminhnguyenai/personal-blog/services/cms/runner/lexer"
)

type writer struct {
	data string
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.data += string(p)
	return len([]byte(w.data)), nil
}

// Return the value of the writer and reset it
func (w *writer) String() string {
	cpy := w.data
	w.data = ""
	return cpy
}

type Renderer struct {
	templates *template.Template
	writer    *writer
}

func NewRenderer() (*Renderer, error) {
	templates, err := template.ParseFiles("runner/renderer/templates.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{templates, &writer{}}, nil
}

func (r *Renderer) Traverse(node *lexer.Node) (string, string) {
	values := ""
	children := ""

	if node.Self.Kind == lexer.FRONTMATTER {
		values += r.frontmatterRenderer(node)
	}

	for _, value := range node.Values {
		switch value.Self.Kind {
		case lexer.PARAGRAPH:
			values += r.inlineParagraphRenderer(value)
		case lexer.LINK:
			values += r.linkRenderer(value)
		case lexer.INLINE_CODE:
			values += r.inlineCodeRenderer(value)
		}
	}

	for _, child := range node.Children {
		switch child.Self.Kind {
		case lexer.HEADING_1, lexer.HEADING_2, lexer.HEADING_3, lexer.HEADING_4, lexer.HEADING_5:
			children += r.headingRenderer(child)
		case lexer.PARAGRAPH:
			children += r.paragraphRenderer(child)
		case lexer.HYPHEN_LIST, lexer.NUMBERED_LIST:
			children += r.listRenderer(child)
		case lexer.CODE_BLOCK:
			children += r.codeBlockRenderer(child)
		case lexer.QUOTE:
			children += r.quoteRenderer(child)
		case lexer.CALLOUT_NOTE, lexer.CALLOUT_IMPORTANT, lexer.CALLOUT_WARNING, lexer.CALLOUT_EXAMPLE:
			children += r.calloutRenderer(child)
		}
	}

	return values, children
}

func (r *Renderer) frontmatterRenderer(node *lexer.Node) string {
	type Data struct {
		Title string
		Date  string
		Tags  []string
	}

	data := Data{}

	for i := 0; i < len(node.Self.Values); i += 2 {
		propertyName := node.Self.Values[i]
		propertyValue := node.Self.Values[i+1]
		switch propertyName {
		case "id":
			data.Title = propertyValue[1 : len(propertyValue)-1]
		case "date":
			data.Date = propertyValue[1 : len(propertyValue)-1]
		case "tags":
			tags := strings.Split(propertyValue, ",")
			data.Tags = tags[:len(tags)-1]
		}
	}

	r.templates.ExecuteTemplate(r.writer, "frontmatter", data)
	return r.writer.String()
}

func (r *Renderer) headingRenderer(node *lexer.Node) string {
	values, children := r.Traverse(node)

	r.templates.ExecuteTemplate(r.writer, "heading", struct {
		Type     int
		Value    template.HTML
		Children template.HTML
	}{int(node.Self.Kind), template.HTML(values), template.HTML(children)})

	return r.writer.String()
}

func (r *Renderer) paragraphRenderer(node *lexer.Node) string {
	r.templates.ExecuteTemplate(r.writer, "paragraph", struct{ Value string }{node.Self.Values[0]})

	return r.writer.String()
}

func (r *Renderer) listRenderer(node *lexer.Node) string {
	values, children := r.Traverse(node)

	if node.Self.Kind == lexer.HYPHEN_LIST {
		r.templates.ExecuteTemplate(r.writer, "hyphen-list", struct {
			Values   template.HTML
			Children template.HTML
		}{template.HTML(values), template.HTML(children)})
	} else if node.Self.Kind == lexer.NUMBERED_LIST {
		r.templates.ExecuteTemplate(r.writer, "numbered-list", struct {
			Number   string
			Values   template.HTML
			Children template.HTML
		}{node.Self.Values[0], template.HTML(values), template.HTML(children)})
	}

	return r.writer.String()
}

func (r *Renderer) quoteRenderer(node *lexer.Node) string {
	_, children := r.Traverse(node)

	r.templates.ExecuteTemplate(r.writer, "quote", struct{ Values template.HTML }{
		template.HTML(children),
	})

	return r.writer.String()
}

func (r *Renderer) calloutRenderer(node *lexer.Node) string {
	values, children := r.Traverse(node)

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

	r.templates.ExecuteTemplate(r.writer, "callout", struct {
		Values   template.HTML
		Children template.HTML
	}{template.HTML(values), template.HTML(children)})

	return r.writer.String()
}

func (r *Renderer) codeBlockRenderer(node *lexer.Node) string {
	r.templates.ExecuteTemplate(r.writer, "codeblock", struct {
		Language string
		Code     string
	}{node.Self.Values[0], node.Self.Values[1]})

	return r.writer.String()
}

func (r *Renderer) inlineParagraphRenderer(node *lexer.Node) string {
	r.templates.ExecuteTemplate(r.writer, "inline-paragraph", struct{ Value string }{node.Self.Values[0]})

	return r.writer.String()
}

func (r *Renderer) linkRenderer(node *lexer.Node) string {
	r.templates.ExecuteTemplate(r.writer, "link", struct {
		Link        string
		Placeholder string
	}{node.Self.Values[1], node.Self.Values[0]})

	return r.writer.String()
}

func (r *Renderer) inlineCodeRenderer(node *lexer.Node) string {
	r.templates.ExecuteTemplate(r.writer, "inline-code", struct{ Value string }{node.Self.Values[0]})

	return r.writer.String()
}

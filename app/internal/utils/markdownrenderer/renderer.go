package markdownrenderer

import (
	"html/template"
	"regexp"
	"strings"

	"github.com/leminhohoho/personal-blog/pkgs/markdownparser"
)

type Writer struct {
	data string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.data += string(p)
	return len([]byte(w.data)), nil
}

// Return the value of the writer and reset it
func (w *Writer) String() string {
	cpy := w.data
	w.data = ""
	return cpy
}

type Renderer struct {
	templates *template.Template
	writer    *Writer
}

func NewRenderer() (*Renderer, error) {
	templates, err := template.New("content").Funcs(FuncsMap).ParseFiles("templates/templates.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{templates, &Writer{}}, nil
}

func (r *Renderer) Render(astTree *markdownparser.Node) string {
	values, children := r.traverse(astTree)
	content := values + children

	return content
}

func (r *Renderer) generateTOC(node *markdownparser.Node) string {
	children := ""

	for _, child := range node.Children {
		switch child.Self.Kind {
		case markdownparser.HEADING_1,
			markdownparser.HEADING_2,
			markdownparser.HEADING_3,
			markdownparser.HEADING_4,
			markdownparser.HEADING_5:
			children += r.tocRenderer(child)
		}
	}

	return children
}

// Get the string representation of the value in plain text
func getStringRepresentation(node *markdownparser.Node) string {
	values := ""
	for _, value := range node.Values {
		values += value.Self.Values[0]
	}

	return strings.ToLower(values)
}

func (r *Renderer) tocRenderer(node *markdownparser.Node) string {
	children := r.generateTOC(node)
	strRepresentation := getStringRepresentation(node)
	values, _ := r.traverse(node)
	link := strings.Replace(strRepresentation, " ", "-", -1)

	r.templates.ExecuteTemplate(r.writer, "TOC", struct {
		Type     int
		Value    template.HTML
		Link     string
		Children template.HTML
	}{int(node.Self.Kind), template.HTML(values), link, template.HTML(children)})

	switch node.Parent.Self.Kind {
	case markdownparser.HEADING_1,
		markdownparser.HEADING_2,
		markdownparser.HEADING_3,
		markdownparser.HEADING_4,
		markdownparser.HEADING_5:
		return r.writer.String()
	default:
		return "<ul>" + r.writer.String() + "</ul>"
	}
}

func (r *Renderer) traverse(node *markdownparser.Node) (string, string) {
	values := ""
	children := ""

	if node.Self.Kind == markdownparser.FRONTMATTER {
		values += r.frontmatterRenderer(node)
	}

	for _, value := range node.Values {
		switch value.Self.Kind {
		case markdownparser.TEXT:
			values += r.defaultRenderer(value, "text")
		case markdownparser.LINK:
			values += r.linkRenderer(value)
		case markdownparser.INLINE_CODE:
			values += r.defaultRenderer(value, "inline-code")
		case markdownparser.BOLD_TEXT:
			values += r.defaultRenderer(value, "bold-text")
		case markdownparser.ITALIC_TEXT:
			values += r.defaultRenderer(value, "italic-text")
		}
	}

	for _, child := range node.Children {
		switch child.Self.Kind {
		case markdownparser.HEADING_1,
			markdownparser.HEADING_2,
			markdownparser.HEADING_3,
			markdownparser.HEADING_4,
			markdownparser.HEADING_5:
			children += r.headingRenderer(child)
		case markdownparser.HYPHEN_LIST, markdownparser.NUMBERED_LIST, markdownparser.PARAGRAPH:
			children += r.listRenderer(child)
		case markdownparser.CODE_BLOCK:
			children += r.codeBlockRenderer(child)
		case markdownparser.QUOTE:
			children += r.quoteRenderer(child)
		case markdownparser.CALLOUT_NOTE,
			markdownparser.CALLOUT_IMPORTANT,
			markdownparser.CALLOUT_WARNING,
			markdownparser.CALLOUT_EXAMPLE:
			children += r.calloutRenderer(child)
		}
	}

	return values, children
}

func (r *Renderer) defaultRenderer(node *markdownparser.Node, template string) string {
	r.templates.ExecuteTemplate(r.writer, template, struct{ Value string }{node.Self.Values[0]})

	return r.writer.String()
}

func (r *Renderer) frontmatterRenderer(node *markdownparser.Node) string {
	type Data struct {
		Title string
		Date  string
		Tags  []string
		TOC   template.HTML
	}

	data := Data{}

	for i := 0; i < len(node.Self.Values); i += 2 {
		propertyName := node.Self.Values[i]
		propertyValue := node.Self.Values[i+1]
		switch propertyName {
		case "id":
			data.Title = propertyValue
		case "date":
			data.Date = propertyValue
		case "tags":
			tags := strings.Split(propertyValue, ",")
			data.Tags = tags[:len(tags)-1]
		}
	}

	data.TOC = template.HTML(r.generateTOC(node))

	r.templates.ExecuteTemplate(r.writer, "frontmatter", data)
	return r.writer.String()
}

func (r *Renderer) headingRenderer(node *markdownparser.Node) string {
	values, children := r.traverse(node)
	strRepresentation := getStringRepresentation(node)
	link := strings.Replace(strRepresentation, " ", "-", -1)

	r.templates.ExecuteTemplate(r.writer, "heading", struct {
		Type     int
		Value    template.HTML
		Link     string
		Children template.HTML
	}{int(node.Self.Kind), template.HTML(values), link, template.HTML(children)})

	return r.writer.String()
}

// FIX: Adding approprite <ul> tags
func (r *Renderer) listRenderer(node *markdownparser.Node) string {
	values, children := r.traverse(node)

	if node.Self.Kind == markdownparser.HYPHEN_LIST {
		r.templates.ExecuteTemplate(r.writer, "hyphen-list", struct {
			Values   template.HTML
			Children template.HTML
		}{template.HTML(values), template.HTML(children)})
	} else if node.Self.Kind == markdownparser.NUMBERED_LIST {
		r.templates.ExecuteTemplate(r.writer, "numbered-list", struct {
			Number   string
			Values   template.HTML
			Children template.HTML
		}{node.Self.Values[0], template.HTML(values), template.HTML(children)})
	} else if node.Self.Kind == markdownparser.PARAGRAPH {
		r.templates.ExecuteTemplate(r.writer, "paragraph", struct {
			Number   string
			Values   template.HTML
			Children template.HTML
		}{node.Self.Values[0], template.HTML(values), template.HTML(children)})
	}

	switch node.Parent.Self.Kind {
	case markdownparser.PARAGRAPH, markdownparser.HYPHEN_LIST, markdownparser.NUMBERED_LIST:
		return r.writer.String()
	default:
		return "<ul>" + r.writer.String() + "</ul>"
	}

}

func (r *Renderer) quoteRenderer(node *markdownparser.Node) string {
	_, children := r.traverse(node)

	r.templates.ExecuteTemplate(r.writer, "quote", struct{ Values template.HTML }{
		template.HTML(children),
	})

	return r.writer.String()
}

func (r *Renderer) calloutRenderer(node *markdownparser.Node) string {
	values, children := r.traverse(node)

	if values == "" {
		switch node.Self.Kind {
		case markdownparser.CALLOUT_NOTE:
			values = "Note"
		case markdownparser.CALLOUT_IMPORTANT:
			values = "Important"
		case markdownparser.CALLOUT_WARNING:
			values = "Warning"
		case markdownparser.CALLOUT_EXAMPLE:
			values = "Example"
		}
	}

	r.templates.ExecuteTemplate(r.writer, "callout", struct {
		Values   template.HTML
		Children template.HTML
	}{template.HTML(values), template.HTML(children)})

	return r.writer.String()
}

func (r *Renderer) codeBlockRenderer(node *markdownparser.Node) string {
	code := strings.Split(node.Self.Values[1], "\n")

	r.templates.ExecuteTemplate(r.writer, "codeblock", struct {
		Metadata string
		Code     []string
	}{node.Self.Values[0], code})

	return r.writer.String()
}

// e.g. Youtube = bg image, Reddit = minimal widget + title + overview
func (r *Renderer) linkRenderer(node *markdownparser.Node) string {
	var linkType string
	if regexp.MustCompile(`https://www\.youtube\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Youtube"
		// NOTE: Failed preview render will be consider as empty
	} else if regexp.MustCompile(`https://github\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Github"
	} else if regexp.MustCompile(`https://www\.reddit\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Reddit"
	} else if regexp.MustCompile(`https://pkg\.go\.dev.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Gopkg"
	}

	r.templates.ExecuteTemplate(r.writer, "link", struct {
		Link        string
		Type        string
		Placeholder string
	}{node.Self.Values[1], linkType, node.Self.Values[0]})

	return r.writer.String()
}

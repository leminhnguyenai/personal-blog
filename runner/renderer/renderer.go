package renderer

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/leminhnguyenai/personal-blog/runner/apis"
	"github.com/leminhnguyenai/personal-blog/runner/lexer"
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

func (r *Renderer) GenerateTOC(node *lexer.Node) string {
	children := ""

	for _, child := range node.Children {
		switch child.Self.Kind {
		case lexer.HEADING_1, lexer.HEADING_2, lexer.HEADING_3, lexer.HEADING_4, lexer.HEADING_5:
			children += r.tocRenderer(child)
		}
	}

	return children
}

// Get the string representation of the value in plain text
func getStringRepresentation(node *lexer.Node) string {
	values := ""
	for _, value := range node.Values {
		values += value.Self.Values[0]
	}

	return strings.ToLower(values)
}

func (r *Renderer) tocRenderer(node *lexer.Node) string {
	children := r.GenerateTOC(node)
	strRepresentation := getStringRepresentation(node)
	values, _ := r.Traverse(node)
	link := strings.Replace(strRepresentation, " ", "-", -1)

	r.templates.ExecuteTemplate(r.writer, "TOC", struct {
		Type     int
		Value    template.HTML
		Link     string
		Children template.HTML
	}{int(node.Self.Kind), template.HTML(values), link, template.HTML(children)})

	return r.writer.String()
}

func (r *Renderer) Traverse(node *lexer.Node) (string, string) {
	values := ""
	children := ""

	if node.Self.Kind == lexer.FRONTMATTER {
		values += r.frontmatterRenderer(node)
	}

	for _, value := range node.Values {
		switch value.Self.Kind {
		case lexer.TEXT:
			values += r.defaultRenderer(value, "text")
		case lexer.LINK:
			values += r.linkRenderer(value)
		case lexer.INLINE_CODE:
			values += r.defaultRenderer(value, "inline-code")
		case lexer.BOLD_TEXT:
			values += r.defaultRenderer(value, "bold-text")
		case lexer.ITALIC_TEXT:
			values += r.defaultRenderer(value, "italic-text")
		}
	}

	for _, child := range node.Children {
		switch child.Self.Kind {
		case lexer.HEADING_1, lexer.HEADING_2, lexer.HEADING_3, lexer.HEADING_4, lexer.HEADING_5:
			children += r.headingRenderer(child)
		case lexer.HYPHEN_LIST, lexer.NUMBERED_LIST, lexer.PARAGRAPH:
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

func (r *Renderer) defaultRenderer(node *lexer.Node, template string) string {
	r.templates.ExecuteTemplate(r.writer, template, struct{ Value string }{node.Self.Values[0]})

	return r.writer.String()
}

func (r *Renderer) frontmatterRenderer(node *lexer.Node) string {
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
			data.Title = propertyValue[1 : len(propertyValue)-1]
		case "date":
			data.Date = propertyValue[1 : len(propertyValue)-1]
		case "tags":
			tags := strings.Split(propertyValue, ",")
			data.Tags = tags[:len(tags)-1]
		}
	}

	data.TOC = template.HTML(r.GenerateTOC(node))

	r.templates.ExecuteTemplate(r.writer, "frontmatter", data)
	return r.writer.String()
}

func (r *Renderer) headingRenderer(node *lexer.Node) string {
	values, children := r.Traverse(node)
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
	} else if node.Self.Kind == lexer.PARAGRAPH {
		r.templates.ExecuteTemplate(r.writer, "paragraph", struct {
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
	code := strings.Split(node.Self.Values[1], "\n")

	r.templates.ExecuteTemplate(r.writer, "codeblock", struct {
		Metadata string
		Code     []string
	}{node.Self.Values[0], code})

	return r.writer.String()
}

func (r *Renderer) youtubePreview(urlStr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	type Data struct {
		apis.Snippet
		VideoURL   string
		ChannelURL string
	}

	ytbUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	id := ytbUrl.Query().Get("v")

	ytbData, err := apis.GetYtbData(ctx, id)
	if err != nil {
		return "", err
	}

	data := Data{ytbData, urlStr, fmt.Sprintf("https://www.youtube.com/channel/%s", ytbData.ChannelID)}

	r.templates.ExecuteTemplate(r.writer, "youtube-preview", data)

	return r.writer.String(), nil
}

func (r *Renderer) githubPreview(urlStr string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// NOTE: We can assume that the url has github.com domain name and can precisely trim the domain name
	urlPath := strings.Split(urlStr[18:], "/")
	owner := urlPath[1]
	repo := urlPath[2]

	type Data struct {
		Owner               string
		Repo                string
		Branches            []apis.Branch
		Commit              apis.Commit
		TimeSinceLastCommit string
	}

	type Response struct {
		data  Data
		error error
	}

	respChan := make(chan Response)

	go func() {
		branches, err := apis.GetGithubBranches(ctx, owner, repo)
		if err != nil {
			respChan <- Response{error: err}
		}

		commit, err := apis.GetGithubLatestCommit(ctx, owner, repo)
		if err != nil {
			respChan <- Response{error: err}
		}

		respChan <- Response{Data{Owner: owner, Repo: repo, Branches: branches, Commit: commit}, nil}
	}()

	for {
		select {
		case res := <-respChan:
			if res.error != nil {
				return "", res.error
			}

			commitDate := res.data.Commit.Commit.Committer.Date
			// Time since last commit in minutes
			timeSinceLastCommit := int(time.Now().Sub(commitDate)) / 60000000000

			if timeSinceLastCommit < 60 {
				res.data.TimeSinceLastCommit = fmt.Sprintf(
					"last updated %d minutes ago\n",
					timeSinceLastCommit,
				)
			} else if timeSinceLastCommit < 1440 {
				res.data.TimeSinceLastCommit = fmt.Sprintf("last updated %d hour(s) ago\n", timeSinceLastCommit/60)
			} else if timeSinceLastCommit < 43200 {
				res.data.TimeSinceLastCommit = fmt.Sprintf("last updated %d day(s) ago\n", timeSinceLastCommit/1440)
			} else if timeSinceLastCommit < 518400 {
				res.data.TimeSinceLastCommit = fmt.Sprintf("last updated %d month(s) ago\n", timeSinceLastCommit/43200)
			} else {
				res.data.TimeSinceLastCommit = fmt.Sprintf("last updated %d year(s) ago\n", timeSinceLastCommit/518400)
			}

			r.templates.ExecuteTemplate(r.writer, "github-preview", res.data)

			return r.writer.String(), nil
		}
	}
}

// e.g. Youtube = bg image, Reddit = minimal widget + title + overview
func (r *Renderer) linkRenderer(node *lexer.Node) string {
	var linkType string
	var preview string

	if regexp.MustCompile(`https://www\.youtube\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Youtube"
		// NOTE: Failed preview render will be consider as empty
		preview, _ = r.youtubePreview(node.Self.Values[1])
	} else if regexp.MustCompile(`https://github\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Github"
		preview, _ = r.githubPreview(node.Self.Values[1])
	} else if regexp.MustCompile(`https://www\.reddit\.com.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Reddit"
	} else if regexp.MustCompile(`https://pkg\.go\.dev.*`).FindString(node.Self.Values[1]) != "" {
		linkType = "Gopkg"
	}

	r.templates.ExecuteTemplate(r.writer, "link", struct {
		Link        string
		Type        string
		Placeholder string
		Preview     template.HTML
	}{node.Self.Values[1], linkType, node.Self.Values[0], template.HTML(preview)})

	return r.writer.String()
}

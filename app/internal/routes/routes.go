package routes

import (
	"bytes"
	"compress/gzip"
	"html/template"
	"net/http"
	"time"

	"github.com/leminhohoho/personal-blog/app/internal/models"
	"github.com/leminhohoho/personal-blog/app/internal/utils/markdownrenderer"
	"github.com/leminhohoho/personal-blog/pkgs/simplelog"
)

var hash int = int(time.Now().Unix())

func RegisterPosts(mux *http.ServeMux, blogs []*models.Blog) error {
	for _, blog := range blogs {
		mux.HandleFunc("GET /"+blog.Name, func(w http.ResponseWriter, r *http.Request) {
			templ, err := template.New("").
				Funcs(markdownrenderer.FuncsMap).
				ParseFiles("templates/index.html", "templates/templates.html")
			if err != nil {
				simplelog.Errorf(err)
			}

			var buffer bytes.Buffer

			templ.ExecuteTemplate(&buffer, "index", struct {
				Content template.HTML
				Hash    int
			}{
				Content: template.HTML(blog.HTMLContent),
				Hash:    hash,
			})

			html := buffer.String()

			var b bytes.Buffer

			compressor := gzip.NewWriter(&b)
			compressor.Write([]byte(html))
			compressor.Close()

			w.Header().Add("Content-Type", "text/html")
			w.Header().Add("Content-Encoding", "gzip")
			w.WriteHeader(http.StatusOK)
			w.Write(b.Bytes())
		})
	}

	return nil
}

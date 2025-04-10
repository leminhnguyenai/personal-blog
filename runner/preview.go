package runner

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/leminhnguyenai/personal-blog/runner/asciitree"
	"github.com/leminhnguyenai/personal-blog/runner/lexer"
	"github.com/leminhnguyenai/personal-blog/runner/renderer"
)

var hash int = int(time.Now().Unix())

type Data struct {
	Content template.HTML
	TOC     template.HTML
	Hash    int
}

func Preview(e *Engine, filePath string) error {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", FileServer("static"))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		e.debug("Connected\n")
		w.Header().Add("Cache-Control", "public, max-age=31536000")

		data, err := os.ReadFile(filePath)
		if err != nil {
			HandleError(w, err)
			return
		}

		astTree, err := lexer.ParseAST(string(data))
		if err != nil {
			HandleError(w, err)
			return
		}

		var str string

		astTree.Display(&str, 0)

		e.debug("\n%s", asciitree.GenerateTree(str))

		// Move the whole renderering process to renderer.go
		mdRenderer, err := renderer.NewRenderer()
		if err != nil {
			HandleError(w, err)
			return
		}
		writer := &renderer.Writer{}

		content := mdRenderer.Render(astTree)

		templ, err := template.New("").Funcs(renderer.FuncsMap).ParseFiles("templates/index.html")
		if err != nil {
			HandleError(w, err)
			return
		}

		templ.ExecuteTemplate(writer, "index", Data{
			Content: template.HTML(content),
			Hash:    hash,
		})

		html := writer.String()

		var b bytes.Buffer

		compressor := gzip.NewWriter(&b)
		compressor.Write([]byte(html))
		compressor.Close()

		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())
	})

	srv := &http.Server{Addr: ":3000", Handler: mux}

	port := os.Getenv("PORT")
	fmt.Printf("The server is live on http://localhost:%v\n", port)

	errChan := make(chan error)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-e.ExitChan:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				return err
			}
			return nil
		}
	}
}

// NOTE: All log will be change to fmt when the tool is finished

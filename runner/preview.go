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

func FileServerMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=31536000")
		handler.ServeHTTP(w, r)
	}
}

func Preview(e *Engine, filePath string) error {
	_, err := GetFreePort()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.Handle("GET /static/", FileServerMiddleware(FileServer("static")))
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

		logData := fmt.Sprintf(`
Time: %v

%s

        `, time.Now(), asciitree.GenerateTree(str))

		if err = Logging(logData, "app.log"); err != nil {
			HandleError(w, err)
			return
		}

		markdownRenderer, err := renderer.NewRenderer()
		if err != nil {
			HandleError(w, err)
			return
		}

		toc := markdownRenderer.GenerateTOC(astTree)
		values, children := markdownRenderer.Traverse(astTree)
		content := values + children

		writer := &renderer.Writer{}

		templ, err := template.New("").Funcs(renderer.FuncsMap).ParseFiles("templates/index.html")
		if err != nil {
			HandleError(w, err)
			return
		}

		templ.ExecuteTemplate(writer, "index", Data{
			Content: template.HTML(content),
			TOC:     template.HTML(toc),
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

	// COMMIT: Create a mechanism to randomly assign a free port if none is selected
	fmt.Printf("The server is live on http://localhost%s\n", ":3000")

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
		case _, ok := <-e.ExitChan:
			if !ok {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				defer cancel()

				if err := srv.Shutdown(ctx); err != nil {
					return err
				}
				return nil
			}
		}
	}
}

// NOTE: All log will be change to fmt when the tool is finished

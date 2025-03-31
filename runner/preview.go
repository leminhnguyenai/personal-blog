package runner

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// COMMIT: Generate different random name for css and js files to enable re-caching
// COMMIT: Create a logging file to save the results and only print out important information
func Preview(filePath string) error {
	_, err := GetFreePort()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	mux.Handle("GET /static/", FileServerMiddleware(FileServer("static")))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Connected")
		w.Header().Add("Cache-Control", "public, max-age=31536000")

		log.Println(r.Header.Get("Accept-Encoding"))

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

		// fmt.Println(asciitree.GenerateTree(str))

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

	errChan := make(chan error)

	go func() {

		fmt.Printf("The server is live on http://localhost%s\n", ":3000")
		// Since ListenAndServe() don't return nil err, we can exclude server close error to handle it ourselves
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	// Send incoming signal to sigChan, but only when SIGINT and SYSTEM signals are triggered
	// Technically using an unbuffered channel will work, but within the signal.Notify func,
	// codes after the sending to the sigChan won't be executed -> signal.Notify won't be fully executed
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			if err := srv.Close(); err != nil {
				return err
			}

			fmt.Println("Bye")

			return nil
		case err := <-errChan:
			return err
		}
	}
}

// NOTE: All log will be change to fmt when the tool is finished

package runner

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/leminhnguyenai/personal-blog/runner/asciitree"
	"github.com/leminhnguyenai/personal-blog/runner/lexer"
	"github.com/leminhnguyenai/personal-blog/runner/renderer"
)

type Data struct {
	Content template.HTML
	TOC     template.HTML
}

func Preview(filePath string) error {
	_, err := GetFreePort()
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("./static"))

	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Connected")
		log.Println(os.Getenv("YOUTUBE_API_KEY"))

		data, err := os.ReadFile(filePath)
		if err != nil {
			HandleError(w, err)
			return
		}

		sourceNode, err := lexer.ParseAST(string(data))
		if err != nil {
			HandleError(w, err)
			return
		}

		var str string

		sourceNode.Display(&str, 0)

		fmt.Println(asciitree.GenerateTree(str))

		markdownRenderer, err := renderer.NewRenderer()
		if err != nil {
			HandleError(w, err)
			return
		}

		toc := markdownRenderer.GenerateTOC(sourceNode)
		values, children := markdownRenderer.Traverse(sourceNode)
		html := values + children

		templ, err := template.ParseFiles("index.html")
		if err != nil {
			HandleError(w, err)
			return
		}

		templ.ExecuteTemplate(w, "index", Data{
			Content: template.HTML(html),
			TOC:     template.HTML(toc),
		})
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

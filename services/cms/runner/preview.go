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

	"github.com/leminhnguyenai/personal-blog/services/cms/runner/asciitree"
	"github.com/leminhnguyenai/personal-blog/services/cms/runner/lexer"
)

type Data struct {
	Content template.HTML
}

func Preview(filePath string) error {
	_, err := GetFreePort()
	if err != nil {
		return err
	}

	srv := &http.Server{Addr: ":3000"}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Connected")
		data, err := os.ReadFile(filePath)
		if err != nil {
			HandleError(w, err)
			return
		}

		frontmatter, err := lexer.ParseAST(string(data))
		if err != nil {
			HandleError(w, err)
			return
		}

		var str string

		frontmatter.Display(&str, 0)

		fmt.Println(asciitree.GenerateTree(str))

		values, children := Traverse(frontmatter)
		html := values + children

		templ, err := template.ParseFiles("index.html")
		if err != nil {
			HandleError(w, err)
			return
		}

		templ.ExecuteTemplate(w, "index", Data{Content: template.HTML(html)})
	})

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

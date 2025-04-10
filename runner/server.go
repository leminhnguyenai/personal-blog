package runner

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/leminhnguyenai/personal-blog/runner/lexer"
	"github.com/leminhnguyenai/personal-blog/runner/renderer"
)

func searchMdFiles(dirPath string) ([]string, error) {
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("The path is not a directory\n")
	}

	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	entries, err := dir.ReadDir(0)
	if err != nil {
		return nil, err
	}

	mdFiles := []string{}

	for _, entry := range entries {
		if entry.IsDir() {
			subMdFiles, err := searchMdFiles(path.Join(dirPath, entry.Name()))
			if err != nil {
				return nil, err
			}

			mdFiles = append(mdFiles, subMdFiles...)
			continue
		}

		if path.Ext(entry.Name()) == ".md" {
			mdFiles = append(mdFiles, path.Join(dirPath, entry.Name()))
		}
	}

	return mdFiles, nil
}

func sanitizeFilename(filepath string) (string, error) {
	var sanitizedFilename string

	for _, char := range filepath[:len(filepath)-3] {
		// Remove unsupported char, such as glyph
		if regexp.MustCompile(`[^a-zA-Z0-9_\-\t\f\v ]`).FindStringIndex(string(char)) != nil {
			continue
		}

		// Replace invalid char, such as whitespaces
		if regexp.MustCompile(`[^\S\r\n]`).FindStringIndex(string(char)) != nil {
			sanitizedFilename += "-"
			continue
		}

		sanitizedFilename += string(char)
	}

	if sanitizedFilename == "" {
		return "", fmt.Errorf("Invalid file name at %s\n", filepath)
	}

	return strings.ToLower(sanitizedFilename), nil
}

// COMMIT: Set up basic server that has route for each markdown files
func Server(e *Engine, dirPath string) error {
	hash := int(time.Now().Unix())

	mux := http.NewServeMux()

	mux.Handle("GET /static/", FileServer("static"))

	mdFiles, err := searchMdFiles(dirPath)
	if err != nil {
		return err
	}

	for _, file := range mdFiles {
		filename, _ := sanitizeFilename(path.Base(file))
		e.debug("%s\n", filename)

		mux.Handle(
			fmt.Sprintf("GET /%s/", filename),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := os.ReadFile(file)
				if err != nil {
					HandleError(w, err)
					return
				}

				astTree, err := lexer.ParseAST(string(data))
				if err != nil {
					HandleError(w, err)
					return
				}

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
			}),
		)
	}

	srv := &http.Server{Addr: ":3000", Handler: mux}

	port := os.Getenv("PORT")
	fmt.Printf("The server is live on http://localhost:%v\n", port)

	errChan := make(chan error)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()

	// FIX: Fix name get stripped begin and end pos
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

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

// Sanitize the filename to be valid URL format
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

func Server(e *Engine, dirPath string) error {
	hash := int(time.Now().Unix())

	mux := http.NewServeMux()

	mux.Handle("GET /static/", FileServer("static"))

	mdFiles, err := searchMdFiles(dirPath)
	if err != nil {
		return err
	}

	type Blog struct {
		Link string
		Name string
	}

	var blogs []Blog

	// Generate path for each file
	for _, file := range mdFiles {
		fileUrl, err := sanitizeFilename(path.Base(file))
		if err != nil {
			return err
		}
		url := fmt.Sprintf("GET /%s/", fileUrl)
		e.debug("%s\n", fileUrl)

		blogs = append(blogs, Blog{"/" + fileUrl, path.Base(file)})

		mux.Handle(
			url,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Cache-Control", "public, max-age=31536000")

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

				templ, err := template.New("").
					Funcs(renderer.FuncsMap).
					ParseFiles("templates/index.html", "templates/templates.html")
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

	// COMMIT: Add homepage
	// NOTE: Treat the request for now as full page request, will later
	// add support for AJAX request for each route
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "public, max-age=31536000")

		writer := &renderer.Writer{}

		homepageTempl, err := template.ParseFiles("templates/homepage.html")
		if err != nil {
			HandleError(w, err)
		}

		homepageTempl.ExecuteTemplate(writer, "homepage", blogs)

		homepage := writer.String()

		templ, err := template.New("").Funcs(renderer.FuncsMap).ParseFiles(
			"templates/index.html",
			"templates/templates.html",
		)
		if err != nil {
			HandleError(w, err)
		}

		templ.ExecuteTemplate(writer, "index", struct {
			Content template.HTML
			Hash    int
		}{template.HTML(homepage), hash})

		html := writer.String()
		e.debug(html)

		var b bytes.Buffer

		compressor := gzip.NewWriter(&b)
		compressor.Write([]byte(html))
		compressor.Close()

		w.Header().Add("Content-Type", "text/html")
		w.Header().Add("Content-Encoding", "gzip")
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())

	}))

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

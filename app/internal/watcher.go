package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/leminhohoho/personal-blog/app/internal/models"
	"github.com/leminhohoho/personal-blog/app/internal/repositories"
	"github.com/leminhohoho/personal-blog/app/internal/utils/markdownrenderer"
	"github.com/leminhohoho/personal-blog/pkgs/filewatcher"
	"github.com/leminhohoho/personal-blog/pkgs/markdownparser"
	"github.com/leminhohoho/personal-blog/pkgs/simplelog"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteDB      = "../sql.db"
	shortDateForm = "2006-Jan-02"
)

// Watcher is responsible for handling file changes and update the database
// accordingly. Watcher run parallel to the server and has no relation to it
type Watcher struct {
	fw       *filewatcher.FileWatcher
	rootPath string

	blogRepo models.BlogRepository
}

func NewWatcher(dirPath string, debugMode bool) (*Watcher, error) {
	sqlite, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile("../schema.sql")
	if err != nil {
		return nil, err
	}

	schema := string(data)

	_, err = sqlite.Exec(
		"drop table if exists tags;drop table if exists posts;drop table if exists posts_tags;",
	)
	if err != nil {
		return nil, err
	}

	_, err = sqlite.Exec(schema)
	if err != nil {
		return nil, err
	}

	db := repositories.NewSQLiteBlogRepository(sqlite)

	fw := filewatcher.NewFileWatcher(dirPath, time.Millisecond*100)

	return &Watcher{fw: fw, rootPath: dirPath, blogRepo: db}, nil
}

func (w *Watcher) Start() {
	if err := w.uploadToDB(); err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			select {
			case ev, ok := <-w.fw.Events:
				if !ok {
					log.Fatalf("File watcher stop unexpectedly\n")
				}

				simplelog.Infof(ev.String())
				if err := w.handler(ev); err != nil {
					simplelog.Errorf(err)
				}
			case err, ok := <-w.fw.Errors:
				if !ok {
					log.Fatalf("File watcher stop unexpectedly\n")
				}

				simplelog.Errorf(err)
			}
		}
	}()

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

func (w *Watcher) uploadToDB() error {

	watchList := slices.DeleteFunc(w.fw.GetWatchList(), func(fileToWatch filewatcher.FileToWatch) bool {
		return path.Ext(fileToWatch.Path) != ".md"
	})

	for _, file := range watchList {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
		defer cancel()

		simplelog.Debugf("Rendering " + file.Path + "\n")

		filename, err := sanitizeFilename(path.Base(file.Path))
		if err != nil {
			return err
		}

		data, err := os.ReadFile(file.Path)
		if err != nil {
			return err
		}

		astTree, err := markdownparser.ParseAST(string(data))
		if err != nil {
			return err
		}

		mdRenderer, err := markdownrenderer.NewRenderer()
		if err != nil {
			return err
		}

		content := mdRenderer.Render(astTree)

		if err = w.blogRepo.AddPost(ctx, models.Blog{
			Name:        filename,
			Path:        file.Path,
			HTMLContent: content,
			ModTime:     file.ModTime.Format(shortDateForm),
		}); err != nil {
			return err
		}
	}

	return nil
}

func (w *Watcher) handler(ev filewatcher.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	filename, err := sanitizeFilename(path.Base(ev.Path()))
	if err != nil {
		return err
	}

	if ev.Deleted() {
		return w.blogRepo.DeletePost(ctx, filename)
	}

	data, err := os.ReadFile(ev.Path())
	if err != nil {
		return err
	}

	astTree, err := markdownparser.ParseAST(string(data))
	if err != nil {
		return err
	}

	mdRenderer, err := markdownrenderer.NewRenderer()
	if err != nil {
		return err
	}

	content := mdRenderer.Render(astTree)

	if ev.Created() {
		return w.blogRepo.AddPost(ctx, models.Blog{
			Name:        filename,
			Path:        ev.Path(),
			HTMLContent: content,
			ModTime:     ev.ModTime().Format(shortDateForm),
		})
	}

	if ev.Modified() {
		return w.blogRepo.UpdatePost(ctx, models.Blog{
			Name:        filename,
			Path:        ev.Path(),
			HTMLContent: content,
			ModTime:     ev.ModTime().Format(shortDateForm),
		})
	}

	return nil
}

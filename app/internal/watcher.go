package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"time"

	"github.com/leminhohoho/personal-blog/app/internal/utils/markdownrenderer"
	"github.com/leminhohoho/personal-blog/pkgs/filewatcher"
	"github.com/leminhohoho/personal-blog/pkgs/markdownparser"
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqliteDB = "../sql.db"
)

// Watcher is responsible for handling file changes and update the database
// accordingly. Watcher run parallel to the server and has no relation to it
type Watcher struct {
	fw       *filewatcher.FileWatcher
	rootPath string

	db *sql.DB
}

func NewWatcher(dirPath string) (*Watcher, error) {
	db, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		return nil, err
	}

	fw := filewatcher.NewFileWatcher(dirPath, time.Millisecond*100)

	return &Watcher{fw: fw, rootPath: dirPath, db: db}, nil
}

func (w *Watcher) Start() {
	if err := w.uploadToDB(); err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case ev, ok := <-w.fw.Events:
			if !ok {
				log.Fatalf("File watcher stop unexpectedly\n")
			}

			log.Println(ev)
		case err, ok := <-w.fw.Errors:
			if !ok {
				log.Fatalf("File watcher stop unexpectedly\n")
			}

			log.Fatal(err)
		}
	}
}

func (w *Watcher) uploadToDB() error {
	data, err := os.ReadFile("../schema.sql")
	if err != nil {
		return err
	}

	schema := string(data)

	_, err = w.db.Exec(
		"drop table if exists tags;drop table if exists posts;drop table if exists posts_tags;",
	)
	if err != nil {
		return err
	}

	_, err = w.db.Exec(schema)
	if err != nil {
		return err
	}

	watchList := slices.DeleteFunc(w.fw.GetWatchList(), func(fileToWatch filewatcher.FileToWatch) bool {
		return path.Ext(fileToWatch.Path) != ".md"
	})

	for _, file := range watchList {
		fmt.Println(file.Path)
		filename := path.Base(file.Path)
		filename = filename[:len(filename)-4]
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

		_, err = w.db.Exec(
			"INSERT INTO posts(file_path, name, content) VALUES(?,?,?)",
			file.Path,
			filename,
			content,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

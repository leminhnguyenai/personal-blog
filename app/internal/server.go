package internal

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/leminhohoho/personal-blog/app/internal/helpers"
	"github.com/leminhohoho/personal-blog/app/internal/middlewares"
	"github.com/leminhohoho/personal-blog/app/internal/models"
	"github.com/leminhohoho/personal-blog/app/internal/repositories"
	"github.com/leminhohoho/personal-blog/app/internal/routes"
	"github.com/leminhohoho/personal-blog/pkgs/simplelog"
)

type Server struct {
	mux *http.ServeMux

	blogRepo models.BlogRepository
}

func NewServer(debugMode bool) (*Server, error) {
	sqlite, err := sql.Open("sqlite3", sqliteDB)
	if err != nil {
		return nil, err
	}

	db := repositories.NewSQLiteBlogRepository(sqlite)

	return &Server{
		mux:      http.NewServeMux(),
		blogRepo: db,
	}, nil
}

func (srv *Server) Construct() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	srv.mux.Handle("GET /static/", helpers.FileServer("static"))

	blogs, err := srv.blogRepo.GetAllPosts(ctx)
	if err != nil {
		return err
	}

	if err := routes.RegisterPosts(srv.mux, blogs); err != nil {
		return err
	}

	// Handlers for HTTP server
	srv.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		simplelog.Debugf(fmt.Sprintf("ContentLength: %d bytes\n", r.ContentLength))
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		simplelog.Debugf(fmt.Sprintf("ContentLength: %d bytes\n", r.ContentLength))
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("PATCH /", func(w http.ResponseWriter, r *http.Request) {
		simplelog.Debugf(fmt.Sprintf("ContentLength: %d bytes\n", r.ContentLength))
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("DELETE /", func(w http.ResponseWriter, r *http.Request) {
		simplelog.Debugf(fmt.Sprintf("ContentLength: %d bytes\n", r.ContentLength))
		w.Write([]byte("Leminhohhoho's blog"))
	})

	return nil
}

func (srv *Server) Start() {
	port := os.Getenv("PORT")

	fmt.Printf("The server is live on http://localhost:%s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), middlewares.LoggerMiddleware(srv.mux)); err != nil {
		log.Fatal(err)
	}
}

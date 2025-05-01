package internal

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/leminhohoho/personal-blog/app/internal/common/logger"
	"github.com/leminhohoho/personal-blog/app/internal/middlewares"
)

type Server struct {
	mux *http.ServeMux

	log *logger.Logger

	// TODO: Add DB
}

func NewServer(debugMode bool) (*Server, error) {
	return &Server{
		mux: http.NewServeMux(),
		log: logger.NewLogger(debugMode),
	}, nil
}

func (srv *Server) Construct() error {
	// Handlers for HTTP server
	srv.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		srv.log.Debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		srv.log.Debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("PATCH /", func(w http.ResponseWriter, r *http.Request) {
		srv.log.Debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("DELETE /", func(w http.ResponseWriter, r *http.Request) {
		srv.log.Debug("ContentLength: %d bytes\n", r.ContentLength)
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

package internal

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/leminhnguyenai/personal-blog/app/internal/common"
	"github.com/leminhnguyenai/personal-blog/app/internal/middlewares"
)

type Server struct {
	mux *http.ServeMux

	debugMode bool

	// TODO: Add DB
}

func NewServer(debugMode bool) (*Server, error) {
	return &Server{
		mux:       http.NewServeMux(),
		debugMode: debugMode,
	}, nil
}

func (srv *Server) Construct(dirPath string) error {
	// Handlers for HTTP server
	srv.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		srv.debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		srv.debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("PATCH /", func(w http.ResponseWriter, r *http.Request) {
		srv.debug("ContentLength: %d bytes\n", r.ContentLength)
		w.Write([]byte("Leminhohhoho's blog"))
	})
	srv.mux.HandleFunc("DELETE /", func(w http.ResponseWriter, r *http.Request) {
		srv.debug("ContentLength: %d bytes\n", r.ContentLength)
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

func (srv *Server) debug(format string, args ...any) {
	if srv.debugMode {
		fmt.Printf("[" + common.ColorString("DEBUG", common.Bold, common.RedFg) + "]: " +
			fmt.Sprintf(format, args...),
		)
	}
}

package internal

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Server struct {
	handler http.Handler

	debugMode bool

	// TODO: Add DB
}

func NewServer(debugMode bool, dirPath string) (*Server, error) {
	mux := http.NewServeMux()

	// Handlers for HTTP server
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Leminhohhoho's blog"))
	})

	return &Server{
		handler:   mux,
		debugMode: debugMode,
	}, nil
}

func (srv *Server) Start() {
	port := os.Getenv("PORT")

	fmt.Printf("The server is live on http://localhost:%s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), srv.handler); err != nil {
		log.Fatal(err)
	}
}

func (srv *Server) debug(format string, args ...any) {
	if srv.debugMode {
		// NOTE: Replace this with your own implementation of format later
		fmt.Printf(format, args...)
	}
}

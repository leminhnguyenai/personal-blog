package runner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

// Ask the kernel for a free port to use
func GetFreePort() (port string, err error) {
	// Bind the socket to port 0, a random free port will then be selected
	a, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	l, err := net.ListenTCP("tcp", a)
	defer l.Close()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(":%d", l.Addr().(*net.TCPAddr).Port), nil
}

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	log.Printf("Error: %s\n", err.Error())
}

func Preview(filePath string) error {
	_, err := GetFreePort()
	if err != nil {
		return err
	}

	srv := &http.Server{Addr: ":3000"}

	http.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(filePath)
		if err != nil {
			handleError(w, err)
		}

		content := `{{ block "content" . }}` + string(data) + `{{ end }}`

		baseTempl, err := template.ParseFiles("index.html")
		if err != nil {
			handleError(w, err)
		}

		templ, err := baseTempl.Parse(content)
		if err != nil {
			handleError(w, err)
		}

		templ.ExecuteTemplate(w, "index", struct{}{})
	})

	fmt.Printf("The server is live on http://localhost%s\n", ":3000")
	// Since ListenAndServe() don't return nil err, we can exclude server close error to handle it ourselves
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	// Send incoming signal to sigChan, but only when SIGINT and SYSTEM signals are triggered
	// Technically using an unbuffered channel will work, but within the signal.Notify func,
	// codes after the sending to the sigChan won't be executed -> signal.Notify won't be fully executed
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutDownCtx, shutDownRelease := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer shutDownRelease()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		return err
	}

	fmt.Println("")
	log.Println("Bye")

	return nil
}

// COMMIT: Add signal listening to listen for updating files & quitting
// NOTE: All log will be change to fmt when the tool is finished

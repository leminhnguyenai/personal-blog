package runner

import (
	"fmt"
	"net"
	"net/http"
	"os"
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

func Preview(filePath string) error {
	mux := http.NewServeMux()

	port, err := GetFreePort()
	if err != nil {
		return err
	}

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			panic(err)
		}

		fmt.Fprintf(w, string(data))
	})

	srv := &http.Server{Addr: port, Handler: mux}

	fmt.Printf("The server is live on http://localhost%s\n", port)
	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

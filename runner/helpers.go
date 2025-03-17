package runner

import (
	"fmt"
	"log"
	"net"
	"net/http"
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

func HandleError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	log.Printf("Error: %s\n", err.Error())
}

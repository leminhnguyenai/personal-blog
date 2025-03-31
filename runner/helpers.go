package runner

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
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

// NOTE: This is a naive implementation of godotenv, it only support single line values currently
func parseBytes(file *os.File) (map[string]string, error) {
	scanner := bufio.NewScanner(file)
	envMap := map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()

		args := strings.SplitN(line, "=", 2)
		if len(args) != 2 {
			return nil, fmt.Errorf("Error parsing key and value for current line:\n%s\n", line)
		}

		key := regexp.MustCompile(`^[a-zA-Z_]+[a-zA-Z0-9_]*`).FindString(args[0])
		if key == "" {
			return nil, fmt.Errorf("Error parsing key for current line:\n%s\n", line)
		}

		var val string

		if regexp.MustCompile(`^"[^\n]+"$`).FindString(args[1]) != "" {
			val = args[1][1 : len(args[1])-1]
		} else if regexp.MustCompile(`^\S+$`).FindString(args[1]) != "" {
			val = args[1]
		} else {
			return nil, fmt.Errorf("Error parsing value for current lin:\n%s\n", line)
		}

		envMap[key] = val
	}

	return envMap, nil
}

func LoadEnv(envPath string, overload bool) error {
	// Load .env and parse it into a map[string]string
	file, err := os.Open(envPath)
	if err != nil {
		return err
	}
	defer file.Close()

	envMap, err := parseBytes(file)
	if err != nil {
		return err
	}

	// Create a map[string]bool to check if an env is available or not and act based on he overload flag
	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	// Update the env variables
	for key, value := range envMap {
		if !currentEnv[key] || overload {
			err = os.Setenv(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Alternative to http.FileServer that support sending compressed file if the request
// include Accept-Encoding header
func FileServer(dirPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodingHeader := r.Header.Get("Accept-Encoding")
		relativePath := template.JSEscapeString(strings.TrimPrefix(r.URL.Path, "/static"))
		fileExt := path.Ext(relativePath)

		// Remove the hash from the relative path
		splits := strings.Split(relativePath, ".")
		possibleHash := splits[len(splits)-2]
		if regexp.MustCompile(`^[0-9]+$`).FindString(possibleHash) != "" {
			relativePath = strings.Join(splits[:len(splits)-2], ".") + fileExt
		}

		if strings.Contains(relativePath, "../") {
			http.Error(w, "Illegal command ../", http.StatusBadRequest)
		}

		filePath := path.Join(dirPath, relativePath)

		if encodingHeader != "" && strings.Contains(encodingHeader, "gzip") {
			supportedCompressingExt := []string{".css", ".js", ".html"}

			if slices.Contains(supportedCompressingExt, fileExt) {
				filePath += ".gz"
				w.Header().Add("Content-Encoding", "gzip")
			}
		}

		switch fileExt {
		case ".css":
			w.Header().Add("Content-type", "text/css")
		case ".html":
			w.Header().Add("Content-type", "text/html")
		case ".js":
			w.Header().Add("Content-type", "text/javascript")
		case ".woff2":
			w.Header().Add("Content-type", "font/woff2")
		}

		http.ServeFile(w, r, filePath)
	})
}

// Write and append to app.log
func Logging(data, logfilePath string) error {
	f, err := os.OpenFile(logfilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(data)); err != nil {
		return err
	}

	return nil
}

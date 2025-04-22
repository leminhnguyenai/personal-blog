package internal

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

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

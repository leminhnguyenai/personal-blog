package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/leminhnguyenai/personal-blog/runner"
)

func init() {

	if err := runner.LoadEnv(".env", true); err != nil {
		log.Fatal(err)
	}
}

func main() {
	// COMMIT: Convert the os signal management to main.go
	// COMMIT: Convert the flag system to fit the new model
	// NOTE: This is for simplicity, as the program grow there will be more than 1 arguement
	var err error
	cfg, err := runner.NewCfg()
	if err != nil {
		log.Fatal(err)
	}

	errChan := make(chan error)

	go func() {
		if err = runner.Execute(cfg); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			fmt.Println("Bye")
			return
		case err = <-errChan:
			log.Fatalf("%s\nThe application stopped\n", err.Error())
		}
	}
}

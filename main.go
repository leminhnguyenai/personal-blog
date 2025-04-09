package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/leminhnguyenai/personal-blog/runner"
)

func init() {
	if err := runner.LoadEnv(".env", true); err != nil {
		log.Fatal(err)
	}

	// Set port to a random free one if not specified in .env
	if os.Getenv("PORT") == "" {
		a, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			log.Fatal(err)
		}

		l, err := net.ListenTCP("tcp", a)
		defer l.Close()
		if err != nil {
			log.Fatal(err)
		}

		os.Setenv("PORT", strconv.Itoa(l.Addr().(*net.TCPAddr).Port))
	}
}

func main() {
	// NOTE: This is for simplicity, as the program grow there will be more than 1 arguement
	var err error
	errChan := make(chan error)

	engine := runner.NewEngine()

	go func() {
		if err = engine.Execute(os.Args[1:]); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			engine.Stop()
			fmt.Println("Bye")
			os.Exit(0)
		case err = <-errChan:
			log.Fatalf("%s\nThe application stopped\n", err.Error())
		}
	}
}

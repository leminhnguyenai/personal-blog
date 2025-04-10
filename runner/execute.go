package runner

import (
	"flag"
	"fmt"
	"slices"
)

type Engine struct {
	debugMode bool
	ExitChan  chan bool
}

func NewEngine() *Engine {
	return &Engine{
		debugMode: false,
		ExitChan:  make(chan bool),
	}
}

func (e *Engine) parseFlag(args []string) {
	flag.BoolVar(&e.debugMode, "d", false, "Debug mode")
	flag.CommandLine.Parse(args)
}

func (e *Engine) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Not enough arguements\n")
	}

	if !slices.Contains([]string{"server", "preview"}, args[0]) {
		args = append([]string{"server"}, args...)
	}

	// Skip the action arguements
	e.parseFlag(args[1:])

	if len(flag.Args()) < 1 {
		return fmt.Errorf("Too few arguements\n")
	}

	errChan := make(chan error)

	go func() {
		switch args[0] {
		case "server":
			// TODO: Add handler for build the whole blog from a dir
			errChan <- Server(e, flag.Args()[0])
		case "preview":
			// BACKLOG: Add support for multiple files preview ???
			errChan <- Preview(e, flag.Args()[0])
		}
	}()

	for {
		select {
		case err := <-errChan:
			return err
		case <-e.ExitChan:
			return nil
		}
	}
}

func (e *Engine) Stop() {
	e.ExitChan <- true
	close(e.ExitChan)
}

func (e *Engine) log(format string, v ...any) {
	fmt.Printf(format, v...)
}

func (e *Engine) debug(format string, v ...any) {
	fmt.Printf("[DEBUG]: ")
	e.log(format, v...)
}

// NOTE: the 'return nil' statements in the first 3 case is a placeholder to make sure that the function is guaranteed to return an error

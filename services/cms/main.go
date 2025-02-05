package main

import (
	"fmt"

	"github.com/leminhnguyenai/personal-blog/services/cms/runner"
)

func main() {
	// NOTE: This is for simplicity, as the program grow there will be more than 1 arguement
	cfg, err := runner.NewCfg()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	runner.Execute(cfg)
}

// COMMIT: Build a basic server to receive a markdown file and stream it

package main

import (
	"fmt"
	"log"

	"github.com/leminhnguyenai/personal-blog/runner"
)

func main() {
	if err := runner.LoadEnv(".env", true); err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	// NOTE: This is for simplicity, as the program grow there will be more than 1 arguement
	cfg, err := runner.NewCfg()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}

	if err = runner.Execute(cfg); err != nil {
		log.Printf("Error: %s\n", err.Error())
	}
}

package runner

import "fmt"

func Execute(cfg Config) {
	switch cfg.Action {
	case ActionCreate:
		// do sumthing
	case ActionUpdate:
		// do sumthing
	case ActionDelete:
		// do sumthing
	case ActionPreview:
		if err := Preview(cfg.FilePath); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}
}

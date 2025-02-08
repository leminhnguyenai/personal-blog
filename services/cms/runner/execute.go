package runner

import "fmt"

func Execute(cfg Config) error {
	switch cfg.Action {
	case ActionCreate:
		// do sumthing
		return nil
	case ActionUpdate:
		// do sumthing
		return nil
	case ActionDelete:
		// do sumthing
		return nil
	case ActionPreview:
		return Preview(cfg.FilePath)
	default:
		return fmt.Errorf("Error reading config")
	}
}

// NOTE: the 'return nil' statements in the first 3 case is a placeholder to make sure that the function is guaranteed to return an error

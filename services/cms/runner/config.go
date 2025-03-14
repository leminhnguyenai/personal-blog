package runner

import (
	"flag"
	"fmt"
)

const (
	ActionCreate  string = "create"
	ActionUpdate  string = "update"
	ActionDelete  string = "delete"
	ActionPreview string = "preview"
)

type Config struct {
	FilePath string
	PostURL  string
	Action   string
}

func NewCfg() (Config, error) {
	// newFilePtr := flag.String("n", "", "Add new document")
	// updateFilePtr := flag.String("m", "", "update document")
	// deleteFilePtr := flag.String("m", "", "delete document")
	previewFilePtr := flag.String("p", "", "preview document")

	flag.Parse()

	if *previewFilePtr != "" {
		return Config{
			FilePath: *previewFilePtr,
			Action:   ActionPreview,
		}, nil
	}

	return Config{}, fmt.Errorf("Error parsing arguements")
}

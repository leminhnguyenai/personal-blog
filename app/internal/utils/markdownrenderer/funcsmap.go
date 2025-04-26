package markdownrenderer

import (
	"html/template"
	"regexp"
)

func sum(a, b int) int {
	return a + b
}

func arr(args ...any) []any {
	return args
}

// Match the filename's file type extension/file type with appropriate filetype
func matchMetadata(metadata string, filetype string) bool {
	pattern := `^[a-zA-Z\.-_]+\.` + filetype + `$`

	if regexp.MustCompile(pattern).FindString(metadata) == "" && metadata != filetype {
		return false
	}

	return true
}

func capitalizeFilename(filename string) string {
	if regexp.MustCompile(`^[a-z]+$`).FindString(filename) == "" {
		return ""
	}

	data := []byte(filename)

	data[0] -= 32

	return string(data)
}

var FuncsMap = template.FuncMap{
	"sum":                sum,
	"arr":                arr,
	"matchMetadata":      matchMetadata,
	"capitalizeFilename": capitalizeFilename,
}

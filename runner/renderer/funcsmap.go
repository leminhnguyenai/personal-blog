package renderer

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

func matchMetadata(metadata string, filetype string) bool {
	pattern := `^[a-zA-Z\.-_]+\.` + filetype + `$`

	if regexp.MustCompile(pattern).FindString(metadata) == "" && metadata != filetype {
		return false
	}

	return true
}

var FuncsMap = template.FuncMap{
	"sum":           sum,
	"arr":           arr,
	"matchMetadata": matchMetadata,
}

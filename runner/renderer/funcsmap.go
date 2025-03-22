package renderer

import "html/template"

var FuncsMap = template.FuncMap{
	"sum": func(a, b int) int {
		return a + b
	},
	"arr": func(args ...string) []string {
		return args
	},
}

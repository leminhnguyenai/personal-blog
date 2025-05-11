package middlewares

import (
	"fmt"
	"net/http"

	"github.com/leminhohoho/personal-blog/pkgs/simplelog"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: Add color to the logger
		switch r.Method {
		case "GET":
			simplelog.Infof(
				fmt.Sprintf(simplelog.ColorString("%s: ", simplelog.Bold+simplelog.GreenFg), r.Method) +
					r.URL.Path + "\n",
			)
		case "POST":
			simplelog.Infof(
				fmt.Sprintf(simplelog.ColorString("%s: ", simplelog.Bold+simplelog.YellowFg), r.Method) +
					r.URL.Path + "\n",
			)
		case "PATCH":
			simplelog.Infof(
				fmt.Sprintf(simplelog.ColorString("%s: ", simplelog.Bold+simplelog.BlueFg), r.Method) +
					r.URL.Path + "\n",
			)
		case "DELETE":
			simplelog.Infof(
				fmt.Sprintf(simplelog.ColorString("%s: ", simplelog.Bold+simplelog.RedFg), r.Method) +
					r.URL.Path + "\n",
			)
		}

		h.ServeHTTP(w, r)
	})
}

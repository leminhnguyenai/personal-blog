package middlewares

import (
	"log"
	"net/http"

	"github.com/leminhohoho/personal-blog/app/internal/common"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: Add color to the logger
		switch r.Method {
		case "GET":
			log.Printf(
				common.ColorString("%s:", common.Bold, common.GreenFg)+"%s\n",
				r.Method,
				r.URL.Path,
			)
		case "POST":
			log.Printf(common.ColorString("%s:", common.Bold, common.YellowFg)+"%s\n", r.Method, r.URL.Path)
		case "PATCH":
			log.Printf(
				common.ColorString("%s:", common.Bold, common.BlueFg)+"%s\n",
				r.Method,
				r.URL.Path,
			)
		case "DELETE":
			log.Printf(common.ColorString("%s:", common.Bold, common.RedFg)+"%s\n", r.Method, r.URL.Path)
		}

		h.ServeHTTP(w, r)
	})
}

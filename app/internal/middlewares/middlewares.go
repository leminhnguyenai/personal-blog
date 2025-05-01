package middlewares

import (
	"log"
	"net/http"

	"github.com/leminhohoho/personal-blog/app/internal/common/logger"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: Add color to the logger
		switch r.Method {
		case "GET":
			log.Printf(
				logger.ColorString("%s:", logger.Bold, logger.GreenFg)+"%s\n",
				r.Method,
				r.URL.Path,
			)
		case "POST":
			log.Printf(logger.ColorString("%s:", logger.Bold, logger.YellowFg)+"%s\n", r.Method, r.URL.Path)
		case "PATCH":
			log.Printf(
				logger.ColorString("%s:", logger.Bold, logger.BlueFg)+"%s\n",
				r.Method,
				r.URL.Path,
			)
		case "DELETE":
			log.Printf(logger.ColorString("%s:", logger.Bold, logger.RedFg)+"%s\n", r.Method, r.URL.Path)
		}

		h.ServeHTTP(w, r)
	})
}

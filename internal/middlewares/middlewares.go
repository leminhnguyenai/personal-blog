package middlewares

import (
	"log"
	"net/http"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: Add color to the logger
		log.Printf("%s: %s", r.Method, r.URL.Path)

		h.ServeHTTP(w, r)
	})
}

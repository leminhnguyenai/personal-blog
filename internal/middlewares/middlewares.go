package middlewares

import (
	"log"
	"net/http"

	"github.com/leminhnguyenai/personal-blog/internal/common"
)

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: Add color to the logger
		log.Printf(common.ColorStringFg("%s:", common.RedFg, common.WhiteBg)+"%s\n", r.Method, r.URL.Path)

		h.ServeHTTP(w, r)
	})
}

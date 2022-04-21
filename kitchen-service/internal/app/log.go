package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/w-k-s/McMicroservices/kitchen-service/log"
)

func loggingMiddleware(logger log.Logger) mux.MiddlewareFunc {
	return mux.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := r.WithContext(logger.
				WithFields(map[string]interface{}{
					"TraceId": "TODO",
				}).
				WithContext(r.Context()),
			)
			next.ServeHTTP(w, req)
		})
	})
}

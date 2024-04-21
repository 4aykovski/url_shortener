package middleware

import (
	"log/slog"
	"net/http"
)

func (m *CustomMiddlewares) TaskManager(log *slog.Logger) func(handler http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		log = log.With(
			slog.String("middleware", "swagger"),
		)

		log.Debug("Swagger middleware initialized")

		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")

			handler.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

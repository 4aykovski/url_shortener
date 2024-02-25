package v1

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const (
	authorizationHeader = "Authorization"
)

var (
	ErrInvalidAuthHeader = errors.New("invalid auth header")
)

func (h *Handler) loggerMiddleware(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/loggerMiddleware"),
		)

		log.Info("loggerMiddleware middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", middleware.GetReqID(r.Context())),
			)

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

func (h *Handler) jWTAuthorization() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := h.parseAuthHeader(r)
			if err != nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.UnauthorizedError())
				return
			}

			ctx := context.WithValue(r.Context(), userCtx, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) parseAuthHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get(authorizationHeader)
	if authHeader == "" {
		return "", ErrInvalidAuthHeader
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}

	if len(headerParts[1]) == 0 {
		return "", ErrInvalidAuthHeader
	}

	claims, err := h.tokenManager.Parse(headerParts[1])
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidAuthHeader, err)
	}

	return claims, err
}

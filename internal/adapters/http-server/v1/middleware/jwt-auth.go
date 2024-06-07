package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/4aykovski/url_shortener/pkg/api/response"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
)

const (
	authorizationHeader = "Authorization"
	UserCtx             = "userId"
)

var (
	ErrInvalidAuthHeader = errors.New("invalid auth header")
)

func (m *CustomMiddlewares) JWTAuthorization(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := m.parseAuthHeader(r)
			if err != nil {
				if errors.Is(err, ErrInvalidAuthHeader) {
					log.Info("invalid auth header", slog.String("error", err.Error()))

					render.Status(r, http.StatusUnauthorized)
					render.JSON(w, r, response.UnauthorizedError())
					return
				}

				log.Error("failed to parse auth header", slog.String("error", err.Error()))

				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, response.InternalError())
				return
			}

			log.Debug("auth header parsed - claims", slog.Any("claims", claims))

			id, ok := claims["user_id"]
			if !ok {
				log.Error("invalid jwt claims in auth header")

				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.UnauthorizedError())
				return
			}

			log.Debug("user_id", slog.String("user_id", id.(string)))

			ctx := context.WithValue(r.Context(), UserCtx, id.(string))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *CustomMiddlewares) parseAuthHeader(r *http.Request) (jwt.MapClaims, error) {

	authHeader := r.Header.Get(authorizationHeader)
	if authHeader == "" {
		return nil, ErrInvalidAuthHeader
	}

	headerParts := strings.Split(authHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, ErrInvalidAuthHeader
	}

	if len(headerParts[1]) == 0 {
		return nil, ErrInvalidAuthHeader
	}

	claims, err := m.tokenManager.Parse(headerParts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAuthHeader, err)
	}

	return claims, err
}

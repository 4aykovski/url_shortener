package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/4aykovski/url_shortener/pkg/api/response"
	"github.com/go-chi/render"
)

const (
	authorizationHeader = "Authorization"
	UserCtx             = "userId"
)

var (
	ErrInvalidAuthHeader = errors.New("invalid auth header")
)

func (m *CustomMiddlewares) JWTAuthorization() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := m.parseAuthHeader(r)
			if err != nil {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.UnauthorizedError())
				return
			}

			ctx := context.WithValue(r.Context(), UserCtx, id)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (m *CustomMiddlewares) parseAuthHeader(r *http.Request) (string, error) {

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

	claims, err := m.tokenManager.Parse(headerParts[1])
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidAuthHeader, err)
	}

	return claims, err
}

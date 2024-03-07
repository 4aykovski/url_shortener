package middleware

import tokenManager "github.com/4aykovski/url_shortener/internal/lib/token-manager"

type CustomMiddlewares struct {
	tokenManager tokenManager.TokenManager
}

func New(tokenManager tokenManager.TokenManager) *CustomMiddlewares {
	return &CustomMiddlewares{tokenManager: tokenManager}
}

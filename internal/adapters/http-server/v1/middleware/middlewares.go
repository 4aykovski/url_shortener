package middleware

import tokenManager "github.com/4aykovski/url_shortener/pkg/manager/token"

type CustomMiddlewares struct {
	tokenManager tokenManager.TokenManager
}

func New(tokenManager tokenManager.TokenManager) *CustomMiddlewares {
	return &CustomMiddlewares{tokenManager: tokenManager}
}

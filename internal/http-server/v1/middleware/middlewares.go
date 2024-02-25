package middleware

import tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"

type CustomMiddlewares struct {
	tokenManager tokenManager.TokenManager
}

func New(tokenManager tokenManager.TokenManager) *CustomMiddlewares {
	return &CustomMiddlewares{tokenManager: tokenManager}
}

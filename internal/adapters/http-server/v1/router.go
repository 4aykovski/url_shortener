package v1

import (
	"context"
	"log/slog"

	"github.com/4aykovski/url_shortener/internal/adapters/http-server/v1/handler"
	"github.com/4aykovski/url_shortener/internal/adapters/http-server/v1/middleware"
	"github.com/4aykovski/url_shortener/internal/services"
	tokenManager "github.com/4aykovski/url_shortener/pkg/manager/token"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type authService interface {
	SignUp(ctx context.Context, input services.AuthSignUpInput) error
	SignIn(ctx context.Context, input services.AuthSignInInput) (*tokenManager.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*tokenManager.Tokens, error)
}

type urlService interface {
	SaveURL(ctx context.Context, input services.SaveURLInput) (string, error)
	GetURL(ctx context.Context, input services.GetURLInput) (string, error)
	DeleteURL(ctx context.Context, input services.DeleteURLInput) error
	GetAllUserUrls(ctx context.Context, input services.GetAllUserUrlsInput) (services.GetAllUserUrlsOutput, error)
}

func NewMux(
	log *slog.Logger,
	urlService urlService,
	authService authService,
	tokenManager tokenManager.TokenManager,
) *chi.Mux {
	var (
		mux               = chi.NewMux()
		userHandler       = handler.NewAuthHandler(authService, tokenManager)
		urlHandler        = handler.NewUrlHandler(urlService)
		customMiddlewares = middleware.New(tokenManager)
	)

	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.RealIP)
	mux.Use(customMiddlewares.Logger(log))
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.URLFormat)

	mux.Route("/api/v1", func(r chi.Router) {
		initUrlRoutes(log, r, urlHandler, customMiddlewares)
		initAuthRoutes(log, r, userHandler, customMiddlewares)
	})

	return mux
}

func initUrlRoutes(log *slog.Logger, r chi.Router, h *handler.UrlHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/urls", func(r chi.Router) {
		r.Get("/{alias}", h.Redirect(log))
		r.Group(func(r chi.Router) {
			r.Use(mws.JWTAuthorization(log))
			r.Post("/", h.Save(log))
			r.Get("/", h.GetAllUserUrls(log))
			r.Delete("/{alias}", h.Delete(log))
		})
	})
}

func initAuthRoutes(log *slog.Logger, r chi.Router, h *handler.AuthHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/users", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", h.SignUp(log))
			r.Post("/signin", h.SignIn(log))
			r.Post("/refresh", h.Refresh(log))
			r.Post("/logout", h.Logout(log))
		})
	})
}

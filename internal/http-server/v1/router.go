package v1

import (
	"context"
	"log/slog"

	"github.com/4aykovski/url_shortener/internal/http-server/v1/handler"
	"github.com/4aykovski/url_shortener/internal/http-server/v1/middleware"
	tokenManager "github.com/4aykovski/url_shortener/internal/lib/token-manager"
	"github.com/4aykovski/url_shortener/internal/services"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type UserService interface {
	SignUp(ctx context.Context, input services.UserSignUpInput) error
	SignIn(ctx context.Context, input services.UserSignInInput) (*tokenManager.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*tokenManager.Tokens, error)
}

type UrlRepository interface {
	SaveURL(urlToSave string, alias string) error
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}

func NewMux(
	log *slog.Logger,
	urlRepo UrlRepository,
	userService UserService,
	tokenManager tokenManager.TokenManager,
) *chi.Mux {
	var (
		mux               = chi.NewMux()
		userHandler       = handler.NewUserHandler(userService, tokenManager)
		urlHandler        = handler.NewUrlHandler(urlRepo)
		customMiddlewares = middleware.New(tokenManager)
	)

	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.RealIP)
	mux.Use(customMiddlewares.Logger(log))
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.URLFormat)

	mux.Route("/api/v1", func(r chi.Router) {
		initUrlRoutes(log, r, urlHandler, customMiddlewares)
		initUserRoutes(log, r, userHandler, customMiddlewares)
	})

	return mux
}

func initUrlRoutes(log *slog.Logger, r chi.Router, h *handler.UrlHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/url", func(r chi.Router) {
		r.Get("/{alias}", h.Redirect(log))
		r.Group(func(r chi.Router) {
			r.Use(mws.JWTAuthorization())
			r.Post("/save", h.Save(log))
			r.Delete("/{alias}", h.Delete(log))
		})
	})
}

func initUserRoutes(log *slog.Logger, r chi.Router, h *handler.UserHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/users", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/signup", h.SignUp(log))
			r.Post("/signin", h.SignIn(log))
			r.Post("/refresh", h.Refresh(log))
			r.Post("/logout", h.Logout(log))
		})
	})
}

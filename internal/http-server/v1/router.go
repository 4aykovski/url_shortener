package v1

import (
	"log/slog"

	"github.com/4aykovski/learning/golang/rest/internal/http-server/v1/handlers"
	"github.com/4aykovski/learning/golang/rest/internal/http-server/v1/middleware"
	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func NewMux(
	log *slog.Logger,
	urlRepo handlers.UrlRepository,
	userService handlers.UserService,
	tokenManager tokenManager.TokenManager,
) *chi.Mux {
	var (
		mux               = chi.NewMux()
		userHandler       = handlers.NewUserHandler(userService, tokenManager)
		urlHandler        = handlers.NewUrlHandler(urlRepo)
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

func initUrlRoutes(log *slog.Logger, r chi.Router, h *handlers.UrlHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/url", func(r chi.Router) {
		r.Get("/{alias}", h.UrlRedirect(log))
		r.Group(func(r chi.Router) {
			r.Use(mws.JWTAuthorization())
			r.Post("/save", h.UrlSave(log))
			r.Delete("/{alias}", h.UrlDelete(log))
		})
	})
}

func initUserRoutes(log *slog.Logger, r chi.Router, h *handlers.UserHandler, mws *middleware.CustomMiddlewares) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/signUp", h.UserSignUp(log))
		r.Post("/signIn", h.UserSignIn(log))
		r.Post("/auth/refresh", h.UserRefresh(log))
		r.Group(func(r chi.Router) {
			r.Use(mws.JWTAuthorization())
			r.Post("/signOut", h.UserSignOut(log))
		})
	})
}

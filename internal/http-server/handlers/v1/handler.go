package v1

import (
	"context"
	"log/slog"

	customMiddleware "github.com/4aykovski/learning/golang/rest/internal/http-server/middleware"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name UrlRepository
type UrlRepository interface {
	SaveURL(urlToSave string, alias string) error
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name UserService
type UserService interface {
	SignUp(ctx context.Context, input services.UserSignUpInput) error
}

type Handler struct {
	UrlRepo     UrlRepository
	UserService UserService
}

func New(urlRepo UrlRepository, userService UserService) *Handler {
	return &Handler{
		UrlRepo:     urlRepo,
		UserService: userService,
	}
}

func (h *Handler) InitRoutes(log *slog.Logger, r *chi.Mux) {
	r.Route("/api/v1", func(r chi.Router) {
		h.initUrlRoutes(log, r)
		h.initUserRoutes(log, r)
	})
}

func (h *Handler) InitMiddlewares(log *slog.Logger, r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
}

func (h *Handler) initUrlRoutes(log *slog.Logger, r chi.Router) {
	r.Route("/url", func(r chi.Router) {
		r.Get("/{alias}", h.urlRedirect(log))
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Authorization(log))
			r.Post("/save", h.urlSave(log))
			r.Delete("/{alias}", h.urlDelete(log))
		})
	})
}

func (h *Handler) initUserRoutes(log *slog.Logger, r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/signUp", h.userSignUp(log))
		r.Post("/signIn", h.userSignIn(log))
		r.Post("/auth/refresh", h.userRefresh(log))
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.Authorization(log))
			r.Post("/signOut", h.userSignOut(log))
		})
	})
}

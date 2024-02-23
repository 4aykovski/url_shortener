package v1

import (
	"log/slog"

	customMiddleware "github.com/4aykovski/learning/golang/rest/internal/http-server/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name UrlRepository
type UrlRepository interface {
	SaveURL(urlToSave string, alias string) error
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}

type Handler struct {
	UrlRepo UrlRepository
}

func New(urlRepo UrlRepository) *Handler {
	return &Handler{UrlRepo: urlRepo}
}

func (h *Handler) InitRoutes(log *slog.Logger, r *chi.Mux) {
	h.initUrlRoutes(log, r)
}

func (h *Handler) InitMiddlewares(log *slog.Logger, r *chi.Mux) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger(log))
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
}

func (h *Handler) initUrlRoutes(log *slog.Logger, r *chi.Mux) {
	r.Route("/url", func(r chi.Router) {
		r.Use(customMiddleware.Authorization(log))
		r.Post("/save", h.urlSave(log))
		r.Delete("/{alias}", h.urlDelete(log))
	})
	r.Get("/{alias}", h.urlRedirect(log))
}

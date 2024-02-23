package delete

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/4aykovski/learning/golang/rest/internal/database"
	resp "github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLDeleter interface {
	DeleteURL(string) error
}

func New(log *slog.Logger, deleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("empty alias")

			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		err := deleter.DeleteURL(alias)
		if errors.Is(err, database.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Info("failed to delete url", "alias", alias)

			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("url deleted", "alias", alias)

		responseOK(w, r, alias)
	}
}

type Response struct {
	resp.Response
	Alias string
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

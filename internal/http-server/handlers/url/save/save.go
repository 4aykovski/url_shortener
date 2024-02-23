package save

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/4aykovski/learning/golang/rest/internal/database"
	resp "github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/slogHelper"
	"github.com/4aykovski/learning/golang/rest/internal/lib/random"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=URLRepository
type URLRepository interface {
	SaveURL(urlToSave string, alias string) error
}

// TODO: move to config
const aliasLength = 6

func New(log *slog.Logger, urlSaver URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", slogHelper.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slogHelper.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		if err = urlSaver.SaveURL(req.URL, alias); err != nil {
			if errors.Is(err, database.ErrUrlExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.JSON(w, r, resp.Error("url already exists"))
				return
			}

			log.Error("failed to add url", slogHelper.Err(err))

			render.JSON(w, r, resp.Error("failed to add url"))
			return
		}

		log.Info("url added")

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}

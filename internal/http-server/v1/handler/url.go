package handler

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/slogHelper"
	"github.com/4aykovski/learning/golang/rest/internal/lib/random"
	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name UrlRepository
type UrlRepository interface {
	SaveURL(urlToSave string, alias string) error
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}

type UrlHandler struct {
	UrlRepo UrlRepository

	tokenManager tokenManager.TokenManager
}

func NewUrlHandler(
	UrlRepo UrlRepository,
) *UrlHandler {
	return &UrlHandler{
		UrlRepo: UrlRepo,
	}
}

type UrlSaveInput struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type aliasResponse struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

const aliasLength = 6

func (h *UrlHandler) Save(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.Save"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req UrlSaveInput

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.DecodeError())
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slogHelper.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewRandomString(aliasLength)
		}

		if err = h.UrlRepo.SaveURL(req.URL, alias); err != nil {
			if errors.Is(err, repository.ErrUrlExists) {
				log.Info("url already exists", slog.String("url", req.URL))

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("url already exists"))
				return
			}

			log.Error("failed to add url", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to add url"))
			return
		}

		log.Info("url added")

		responseOK(w, r, alias)
	}
}

func (h *UrlHandler) Redirect(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.Redirect"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("empty alias")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.InvalidRequestError())
			return
		}

		resURL, err := h.UrlRepo.GetURL(alias)
		if errors.Is(err, repository.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}

func (h *UrlHandler) Delete(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.Delete"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("empty alias")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.InvalidRequestError())
			return
		}

		err := h.UrlRepo.DeleteURL(alias)
		if errors.Is(err, repository.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Info("failed to delete url", "alias", alias)

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("url deleted", "alias", alias)

		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, aliasResponse{
		Response: resp.OK(),
		Alias:    alias,
	})
}

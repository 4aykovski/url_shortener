package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/4aykovski/url_shortener/internal/services"
	resp "github.com/4aykovski/url_shortener/pkg/api/response"
	"github.com/4aykovski/url_shortener/pkg/logger/slogHelper"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type urlService interface {
	SaveURL(ctx context.Context, input services.SaveURLInput) (string, error)
	GetURL(ctx context.Context, input services.GetURLInput) (string, error)
	GetAllUserUrls(ctx context.Context, input services.GetAllUserUrlsInput) (services.GetAllUserUrlsOutput, error)
	DeleteURL(ctx context.Context, input services.DeleteURLInput) error
}

type UrlHandler struct {
	urlService urlService
}

func NewUrlHandler(
	urlService urlService,
) *UrlHandler {
	return &UrlHandler{
		urlService: urlService,
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

func (h *UrlHandler) Save(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.Save"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req UrlSaveInput

		userId, ok := getUserId(r.Context())
		if !ok {
			log.Error("failed to get user id")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

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

		alias, err := h.urlService.SaveURL(r.Context(), services.SaveURLInput{
			URL:    req.URL,
			Alias:  req.Alias,
			UserId: userId,
		})
		if err != nil {
			if errors.Is(err, services.ErrAliasAlreadyExists) {
				log.Info("alias already exists", slog.String("alias", req.Alias))

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("alias already exists"))
				return
			}
			log.Error("failed to save url", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("url added")

		responseOK(w, r, alias)
	}
}

type GetAllUserUrlsResponse struct {
	resp.Response
	Urls map[string]string `json:"urls"`
}

func (h *UrlHandler) GetAllUserUrls(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.GetAllUserUrls"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userId, ok := getUserId(r.Context())
		if !ok {
			log.Error("failed to get user id")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		output, err := h.urlService.GetAllUserUrls(r.Context(), services.GetAllUserUrlsInput{
			UserId: userId,
		})
		if err != nil {
			if errors.Is(err, services.ErrUserHasNoUrls) {
				log.Info("user has no urls")

				render.Status(r, http.StatusOK)
				render.JSON(w, r, GetAllUserUrlsResponse{
					Response: resp.OK(),
				})
				return
			}

			log.Error("failed to get urls", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("urls fetched")

		render.Status(r, http.StatusOK)
		render.JSON(w, r, GetAllUserUrlsResponse{
			Response: resp.OK(),
			Urls:     output.Urls,
		})
	}

}

func (h *UrlHandler) Redirect(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.url.Redirect"

		log := log.With(
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

		resURL, err := h.urlService.GetURL(r.Context(), services.GetURLInput{
			Alias: alias,
		})
		if err != nil {
			if errors.Is(err, services.ErrURLNotFound) {
				log.Info("url not found", "alias", alias)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}

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

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		userId, ok := getUserId(r.Context())
		if !ok {
			log.Error("failed to get user id")
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("empty alias")

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.InvalidRequestError())
			return
		}

		err := h.urlService.DeleteURL(r.Context(), services.DeleteURLInput{
			Alias:  alias,
			UserId: userId,
		})
		if err != nil {
			if errors.Is(err, services.ErrURLNotFound) {
				log.Info("url not found", "alias", alias)

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("url not found"))
				return
			}

			log.Error("failed to delete url", slogHelper.Err(err))

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

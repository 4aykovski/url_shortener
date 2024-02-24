package v1

import (
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/slogHelper"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type userSignUpInput struct {
	Login    string `json:"login" validate:"required,min=4,max=128"`
	Password string `json:"password" validate:"required,min=8,max=72,containsany=!*&^?#@)(-+=$_"`
}

func (h *Handler) userSignUp(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.v1.user.SignUp"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req userSignUpInput
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", slogHelper.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded")

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slogHelper.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		if err = h.UserService.SignUp(r.Context(), services.UserSignUpInput{
			Login:    req.Login,
			Password: req.Password,
		}); err != nil {
			if errors.Is(err, repository.ErrUserExists) {
				log.Info("user already exists", slog.String("login", req.Login))

				render.JSON(w, r, resp.Error("Given login is already in use!"))
				return
			}

			log.Error("failed to create user", slogHelper.Err(err))

			render.JSON(w, r, resp.Error("Internal error"))
			return
		}

		log.Info("user created", slog.String("login", req.Login))

		render.JSON(w, r, resp.OK())
	}
}

func (h *Handler) userSignIn(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) userSignOut(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) userRefresh(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

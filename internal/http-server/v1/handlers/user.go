package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	resp "github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/slogHelper"
	tokenManager "github.com/4aykovski/learning/golang/rest/internal/lib/token-manager"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name UserService
type UserService interface {
	SignUp(ctx context.Context, input services.UserSignUpInput) error
	SignIn(ctx context.Context, input services.UserSignInInput) (*services.Tokens, error)
}

type UserHandler struct {
	UserService UserService

	tokenManager tokenManager.TokenManager
}

func NewUserHandler(
	userService UserService,
	tokenManager tokenManager.TokenManager,
) *UserHandler {
	return &UserHandler{
		UserService:  userService,
		tokenManager: tokenManager,
	}
}

type userSignUpInput struct {
	Login    string `json:"login" validate:"required,min=4,max=128"`
	Password string `json:"password" validate:"required,min=8,max=72,containsany=!*&^?#@)(-+=$_"`
}

func (h *UserHandler) UserSignUp(log *slog.Logger) http.HandlerFunc {
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

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.DecodeError())
			return
		}

		log.Info("request body decoded")

		if err = validator.New().Struct(req); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slogHelper.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		if err = h.UserService.SignUp(r.Context(), services.UserSignUpInput{
			Login:    req.Login,
			Password: req.Password,
		}); err != nil {
			if errors.Is(err, repository.ErrUserExists) {
				log.Info("user already exists", slog.String("login", req.Login))

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error("Given login is already in use!"))
				return
			}

			log.Error("failed to create user", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("user created", slog.String("login", req.Login))

		render.JSON(w, r, resp.OK())
	}
}

type userSignInInput struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type tokenResponse struct {
	resp.Response
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *UserHandler) UserSignIn(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.v1.user.userSignIn"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var inp userSignInInput
		err := render.DecodeJSON(r.Body, &inp)
		if err != nil {
			log.Error("can't decode request body", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.DecodeError())
			return
		}

		if err = validator.New().Struct(inp); err != nil {
			var validateErr validator.ValidationErrors
			errors.As(err, &validateErr)

			log.Error("invalid request", slogHelper.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		tokens, err := h.UserService.SignIn(r.Context(), services.UserSignInInput{
			Login:    inp.Login,
			Password: inp.Password,
		})
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) || errors.Is(err, services.ErrWrongCred) {
				log.Info("user not found", slog.String("login", inp.Login))

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.WrongCredentialsError())
				return
			}

			log.Error("can't sign in", slog.String("login", inp.Login), slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		log.Info("successfully signed in", slog.String("login", inp.Login))

		render.JSON(w, r, tokenResponse{
			Response:     resp.OK(),
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
	}
}

func (h *UserHandler) UserSignOut(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *UserHandler) UserRefresh(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

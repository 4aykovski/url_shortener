package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/4aykovski/url_shortener/internal/adapters/repository"
	"github.com/4aykovski/url_shortener/internal/services"
	resp "github.com/4aykovski/url_shortener/pkg/api/response"
	"github.com/4aykovski/url_shortener/pkg/logger/slogHelper"
	tokenManager "github.com/4aykovski/url_shortener/pkg/manager/token"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

const (
	refreshCookieName = "refreshToken"
	refreshCookiePath = "/api/v1/users/auth"
)

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name AuthService
type AuthService interface {
	SignUp(ctx context.Context, input services.AuthSignUpInput) error
	SignIn(ctx context.Context, input services.AuthSignInInput) (*tokenManager.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*tokenManager.Tokens, error)
}

type AuthHandler struct {
	AuthService AuthService

	tokenManager tokenManager.TokenManager
}

func NewAuthHandler(
	authService AuthService,
	tokenManager tokenManager.TokenManager,
) *AuthHandler {
	return &AuthHandler{
		AuthService:  authService,
		tokenManager: tokenManager,
	}
}

type authSignUpInput struct {
	Login    string `json:"login" validate:"required,min=4,max=128"`
	Password string `json:"password" validate:"required,min=8,max=72,containsany=!*&^?#@)(-+=$_"`
}

func (h *AuthHandler) SignUp(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.user.SignUp"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req authSignUpInput

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

		if err = h.AuthService.SignUp(r.Context(), services.AuthSignUpInput{
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

type authSignInInput struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type tokenResponse struct {
	resp.Response
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) SignIn(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.user.SignIn"

		log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("123")

		var inp authSignInInput
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

		tokens, err := h.AuthService.SignIn(r.Context(), services.AuthSignInInput{
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

		// TODO: secure - true в прод
		refreshCookie := h.newRefreshCookie(tokens.RefreshToken, tokens.ExpiresIn)
		http.SetCookie(w, refreshCookie)

		log.Info("successfully signed in", slog.String("login", inp.Login))

		render.JSON(w, r, tokenResponse{
			Response:     resp.OK(),
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
	}
}

type authLogoutInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Logout(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.user.Logout"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var token string
		cookie, err := r.Cookie(refreshCookieName)
		if err != nil {
			var res authLogoutInput
			err = render.DecodeJSON(r.Body, &res)
			if err != nil || res.RefreshToken == "" {
				log.Info("refreshCookie is not specified")

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.WrongCredentialsError())
				return
			}

			token = res.RefreshToken
		} else {
			token = cookie.Value
		}

		err = h.AuthService.Logout(r.Context(), token)
		if err != nil {
			if errors.Is(err, repository.ErrRefreshSessionNotFound) {
				log.Info("can't find session")

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.WrongCredentialsError())
				return
			}

			log.Error("can't logout", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		refreshCookie := h.newRefreshCookie("", time.Unix(0, 0))
		http.SetCookie(w, refreshCookie)

		log.Info("successfully logged out")

		render.JSON(w, r, resp.OK())
	}
}

type authRefreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "v1.handler.user.Refresh"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var token string
		cookie, err := r.Cookie(refreshCookieName)
		if err != nil {
			var res authRefreshInput
			err = render.DecodeJSON(r.Body, &res)
			if err != nil || res.RefreshToken == "" {
				log.Info("refreshCookie is not specified")

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.WrongCredentialsError())
				return
			}
			token = res.RefreshToken
		} else {
			token = cookie.Value
		}

		emptyCookie := h.newRefreshCookie("", time.Now().Add(-100*time.Second))
		http.SetCookie(w, emptyCookie)

		tokens, err := h.AuthService.Refresh(r.Context(), token)
		if err != nil {
			if errors.Is(err, repository.ErrRefreshSessionNotFound) {
				log.Info("can't find session")

				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.WrongCredentialsError())
				return
			}

			log.Error("can't refresh tokens", slogHelper.Err(err))

			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.InternalError())
			return
		}

		// TODO: secure - true в прод
		refreshCookie := h.newRefreshCookie(tokens.RefreshToken, tokens.ExpiresIn)
		http.SetCookie(w, refreshCookie)

		log.Info("successfully refreshed tokens")

		render.JSON(w, r, tokenResponse{
			Response:     resp.OK(),
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
	}
}

func (h *AuthHandler) newRefreshCookie(refreshToken string, time time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Expires:  time,
		Path:     refreshCookiePath,
		Secure:   false,
		HttpOnly: true,
		SameSite: 3,
	}
}

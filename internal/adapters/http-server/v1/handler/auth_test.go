package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4aykovski/url_shortener/internal/adapters/http-server/v1/handler/mocks"
	"github.com/4aykovski/url_shortener/internal/adapters/repository"
	"github.com/4aykovski/url_shortener/internal/services"
	"github.com/4aykovski/url_shortener/pkg/api"
	"github.com/4aykovski/url_shortener/pkg/api/response"
	"github.com/4aykovski/url_shortener/pkg/logger/handlers/slogdiscard"
	tokenManager "github.com/4aykovski/url_shortener/pkg/manager/token"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name      string
		inp       authSignUpInput
		status    string
		respError string
		mockError error
	}{
		{
			name: "success sign up",
			inp: authSignUpInput{
				Login:    "ssff23",
				Password: "qwerty123!",
			},
			status: response.StatusOK,
		},
		{
			name: "invalid login max",
			inp: authSignUpInput{
				Login:    "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be smaller than 128 symbols",
		},
		{
			name: "invalid login min",
			inp: authSignUpInput{
				Login:    "ss",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be longer than 4 symbols",
		},
		{
			name: "invalid password min",
			inp: authSignUpInput{
				Login:    "ssff",
				Password: "q",
			},
			status:    response.StatusError,
			respError: "field Password must be longer than 8 symbols",
		},
		{
			name: "invalid password max",
			inp: authSignUpInput{
				Login:    "ssff",
				Password: "qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!123",
			},
			status:    response.StatusError,
			respError: "field Password must be smaller than 72 symbols",
		},
		{
			name: "invalid password spec character",
			inp: authSignUpInput{
				Login:    "ssff",
				Password: "qwerty123",
			},
			status:    response.StatusError,
			respError: "field Password must contains any of special character",
		},
		{
			name: "invalid password required",
			inp: authSignUpInput{
				Login:    "ssff",
				Password: "",
			},
			status:    response.StatusError,
			respError: "field Password is a required field",
		},
		{
			name: "invalid login required",
			inp: authSignUpInput{
				Login:    "",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login is a required field",
		},
		{
			name: "user already exists",
			inp: authSignUpInput{
				Login:    "ssff24",
				Password: "qwerty123!5",
			},
			status:    response.StatusError,
			respError: "Given login is already in use!",
			mockError: repository.ErrUserExists,
		},
		{
			name: "unexpected error",
			inp: authSignUpInput{
				Login:    "ssff24",
				Password: "qwerty123!5",
			},
			status:    response.StatusError,
			respError: response.InternalErrorMessage,
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userService := mocks.NewUserService(t)

			inp := services.AuthSignUpInput{
				Login:    tc.inp.Login,
				Password: tc.inp.Password,
			}

			if tc.respError == "" || tc.mockError != nil {
				userService.On("SignUp", mock.Anything, inp).
					Return(tc.mockError).Once()
			}

			r := chi.NewRouter()
			h := NewAuthHandler(userService, nil).SignUp(slogdiscard.NewDiscardLogger())
			r.Post("/api/v1/users/signUp", h)

			ts := httptest.NewServer(r)
			defer ts.Close()

			u := fmt.Sprintf(ts.URL + "/api/v1/users/signUp")
			reqBody, err := json.Marshal(inp)
			require.NoError(t, err)

			body, err := api.SendRequest(http.MethodPost, u, bytes.NewReader(reqBody))
			require.NoError(t, err)

			var resp response.Response
			err = json.Unmarshal(body, &resp)
			require.NoError(t, err)
			require.Equal(t, tc.status, resp.Status)
			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

func TestSignInHandler(t *testing.T) {
	tests := []struct {
		name      string
		input     authSignInInput
		tokens    tokenManager.Tokens
		status    string
		respError string
		mockError error
	}{
		{
			name: "success signed in",
			input: authSignInInput{
				Login:    "ssff23",
				Password: "qwerty123!4",
			},
			tokens: tokenManager.Tokens{
				AccessToken:  "random string",
				RefreshToken: "random string",
			},
			status: response.StatusOK,
		},
		{
			name: "invalid credentials",
			input: authSignInInput{
				Login:    "1",
				Password: "1!4",
			},
			tokens:    tokenManager.Tokens{},
			status:    response.StatusError,
			respError: response.WrongCredentialsErrorMessage,
			mockError: repository.ErrUserNotFound,
		},
		{
			name: "unexpected error",
			input: authSignInInput{
				Login:    "ssff23",
				Password: "qwerty123!4",
			},
			tokens:    tokenManager.Tokens{},
			status:    response.StatusError,
			respError: response.InternalErrorMessage,
			mockError: errors.New("unexpected error"),
		},
		{
			name: "empty input",
			input: authSignInInput{
				Login:    "",
				Password: "",
			},
			tokens:    tokenManager.Tokens{},
			status:    response.StatusError,
			respError: "field Login is a required field, field Password is a required field",
		},

		{
			name: "empty login input",
			input: authSignInInput{
				Login:    "",
				Password: "qwerty123!4",
			},
			tokens:    tokenManager.Tokens{},
			status:    response.StatusError,
			respError: "field Login is a required field",
		},
		{
			name: "empty password input",
			input: authSignInInput{
				Login:    "ssff23",
				Password: "",
			},
			tokens:    tokenManager.Tokens{},
			status:    response.StatusError,
			respError: "field Password is a required field",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userService := mocks.NewUserService(t)

			inp := services.AuthSignInInput{Login: tc.input.Login, Password: tc.input.Password}

			if tc.respError == "" || tc.mockError != nil {
				userService.On("SignIn", mock.Anything, inp).
					Return(&tc.tokens, tc.mockError).Once()
			}

			r := chi.NewRouter()
			h := NewAuthHandler(userService, nil).SignIn(slogdiscard.NewDiscardLogger())
			r.Post("/api/v1/users/signIn", h)

			ts := httptest.NewServer(r)
			defer ts.Close()

			u := fmt.Sprintf(ts.URL + "/api/v1/users/signIn")
			reqBody, err := json.Marshal(inp)
			assert.NoError(t, err)

			body, err := api.SendRequest(http.MethodPost, u, bytes.NewReader(reqBody))
			require.NoError(t, err)

			var resp tokenResponse
			err = json.Unmarshal(body, &resp)
			require.NoError(t, err)
			require.Equal(t, tc.status, resp.Status)
			require.Equal(t, tc.respError, resp.Error)
			require.Equal(t, tc.tokens.AccessToken, resp.AccessToken)
			require.Equal(t, tc.tokens.RefreshToken, resp.RefreshToken)
		})
	}
}

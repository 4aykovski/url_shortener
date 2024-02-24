package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/4aykovski/learning/golang/rest/internal/http-server/handlers/v1/mocks"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/handlers/slogdiscard"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name      string
		inp       userSignUpInput
		status    string
		respError string
		mockError error
	}{
		{
			name: "success sign up",
			inp: userSignUpInput{
				Login:    "ssff23",
				Password: "qwerty123!",
			},
			status: response.StatusOK,
		},
		{
			name: "invalid login max",
			inp: userSignUpInput{
				Login:    "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be smaller than 128 symbols",
		},
		{
			name: "invalid login min",
			inp: userSignUpInput{
				Login:    "ss",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be longer than 4 symbols",
		},
		{
			name: "invalid password min",
			inp: userSignUpInput{
				Login:    "ssff",
				Password: "q",
			},
			status:    response.StatusError,
			respError: "field Password must be longer than 8 symbols",
		},
		{
			name: "invalid password max",
			inp: userSignUpInput{
				Login:    "ssff",
				Password: "qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!123",
			},
			status:    response.StatusError,
			respError: "field Password must be smaller than 72 symbols",
		},
		{
			name: "invalid password spec character",
			inp: userSignUpInput{
				Login:    "ssff",
				Password: "qwerty123",
			},
			status:    response.StatusError,
			respError: "field Password must contains any of special character",
		},
		{
			name: "invalid password required",
			inp: userSignUpInput{
				Login:    "ssff",
				Password: "",
			},
			status:    response.StatusError,
			respError: "field Password is a required field",
		},
		{
			name: "invalid login required",
			inp: userSignUpInput{
				Login:    "",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login is a required field",
		},
		{
			name: "user already exists",
			inp: userSignUpInput{
				Login:    "ssff24",
				Password: "qwerty123!5",
			},
			status:    response.StatusError,
			respError: "Given login is already in use!",
			mockError: repository.ErrUserExists,
		},
		{
			name: "unexpected error",
			inp: userSignUpInput{
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

			inp := services.UserSignUpInput{
				Login:    tc.inp.Login,
				Password: tc.inp.Password,
			}

			if tc.respError == "" || tc.mockError != nil {
				userService.On("SignUp", mock.Anything, inp).
					Return(tc.mockError).Once()
			}

			r := chi.NewRouter()
			h := New(nil, userService).userSignUp(slogdiscard.NewDiscardLogger())
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
		input     userSignInInput
		status    string
		respError string
		mockError error
	}{
		{
			name: "success signed in",
			input: userSignInInput{
				Login:    "ssff23",
				Password: "qwerty123!4",
			},
			status: response.StatusOK,
		},
		{
			name: "invalid credentials",
			input: userSignInInput{
				Login:    "1",
				Password: "1!4",
			},
			status:    response.StatusError,
			respError: response.WrongCredentialsErrorMessage,
			mockError: repository.ErrUserNotFound,
		},
		{
			name: "unexpected error",
			input: userSignInInput{
				Login:    "ssff23",
				Password: "qwerty123!4",
			},
			status:    response.StatusError,
			respError: response.InternalErrorMessage,
			mockError: errors.New("unexpected error"),
		},
		{
			name: "empty input",
			input: userSignInInput{
				Login:    "",
				Password: "",
			},
			status:    response.StatusError,
			respError: "field Login is a required field, field Password is a required field",
		},

		{
			name: "empty login input",
			input: userSignInInput{
				Login:    "",
				Password: "qwerty123!4",
			},
			status:    response.StatusError,
			respError: "field Login is a required field",
		},
		{
			name: "empty password input",
			input: userSignInInput{
				Login:    "ssff23",
				Password: "",
			},
			status:    response.StatusError,
			respError: "field Password is a required field",
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userService := mocks.NewUserService(t)

			inp := services.UserSignInInput{Login: tc.input.Login, Password: tc.input.Password}

			if tc.respError == "" || tc.mockError != nil {
				userService.On("SignIn", mock.Anything, inp).
					Return(tc.mockError).Once()
			}

			r := chi.NewRouter()
			h := New(nil, userService).userSignIn(slogdiscard.NewDiscardLogger())
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
		})
	}
}

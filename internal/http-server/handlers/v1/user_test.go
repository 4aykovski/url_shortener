package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/4aykovski/learning/golang/rest/internal/http-server/handlers/v1/mocks"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api/response"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/handlers/slogdiscard"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/4aykovski/learning/golang/rest/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSignUpHandler(t *testing.T) {
	tests := []struct {
		name      string
		inp       services.UserSignUpInput
		status    string
		respError string
		mockError error
	}{
		{
			name: "success sign up",
			inp: services.UserSignUpInput{
				Login:    "ssff23",
				Password: "qwerty123!",
			},
			status: response.StatusOK,
		},
		{
			name: "invalid login max",
			inp: services.UserSignUpInput{
				Login:    "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be smaller than 128 symbols",
		},
		{
			name: "invalid login min",
			inp: services.UserSignUpInput{
				Login:    "ss",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field Login must be longer than 8 symbols",
		},
		{
			name: "invalid password min",
			inp: services.UserSignUpInput{
				Login:    "ssff",
				Password: "q",
			},
			status:    response.StatusError,
			respError: "field Password must be longer than 8 symbols",
		},
		{
			name: "invalid password max",
			inp: services.UserSignUpInput{
				Login:    "ssff",
				Password: "qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!qwerty123!123",
			},
			status:    response.StatusError,
			respError: "field Password must be smaller than 72 symbols",
		},
		{
			name: "invalid password spec character",
			inp: services.UserSignUpInput{
				Login:    "ssff",
				Password: "qwerty123",
			},
			status:    response.StatusError,
			respError: "field Password must contains any of special character",
		},
		{
			name: "invalid password required",
			inp: services.UserSignUpInput{
				Login:    "ssff",
				Password: "",
			},
			status:    response.StatusError,
			respError: "field Password is a required field",
		},
		{
			name: "invalid login required",
			inp: services.UserSignUpInput{
				Login:    "",
				Password: "qwerty123!",
			},
			status:    response.StatusError,
			respError: "field login is a required field",
		},
		{
			name: "user already exists",
			inp: services.UserSignUpInput{
				Login:    "ssff24",
				Password: "qwerty123!5",
			},
			status:    response.StatusError,
			respError: "Given login is already in use!",
			mockError: repository.ErrUserExists,
		},
		{
			name: "unexpected error",
			inp: services.UserSignUpInput{
				Login:    "ssff24",
				Password: "qwerty123!5",
			},
			status:    response.StatusError,
			respError: "Internal error",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userService := mocks.NewUserService(t)

			if tc.respError == "" || tc.mockError != nil {
				userService.On("SignUp", mock.Anything, tc.inp).
					Return(tc.mockError).Once()
			}

			r := chi.NewRouter()
			h := New(nil, userService).userSignUp(slogdiscard.NewDiscardLogger())
			r.Post("/api/v1/users/signUp", h)

			ts := httptest.NewServer(r)
			defer ts.Close()

			u := fmt.Sprintf(ts.URL + "/api/v1/users/signUp")
			body, err := api.SignUpUser(u, tc.inp)
			require.NoError(t, err)

			var resp response.Response
			err = json.Unmarshal(body, &resp)
			require.NoError(t, err)
			require.Equal(t, tc.status, resp.Status)
		})
	}
}

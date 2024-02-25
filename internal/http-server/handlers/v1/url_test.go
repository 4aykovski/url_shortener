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
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/handlers/slogdiscard"
	"github.com/4aykovski/learning/golang/rest/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "success",
			alias: "test_alias",
			url:   "https://google.com/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			urlRepo := mocks.NewUrlRepository(t)

			if tc.respError == "" || tc.mockError != nil {
				urlRepo.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			h := New(urlRepo, nil, nil)
			r := chi.NewRouter()
			r.Get("/{alias}", h.urlRedirect(slogdiscard.NewDiscardLogger()))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToUrl, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			require.Equal(t, tc.url, redirectedToUrl)
		})
	}
}

func TestSaveHandler(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://google.com",
		},
		{
			name:  "Empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:      "Empty url",
			url:       "",
			alias:     "some_alias",
			respError: "field URL is a required field",
		},
		{
			name:      "Invalid url",
			url:       "some invalid url",
			alias:     "some_alias",
			respError: "field URL is not a valid URL",
		},
		{
			name:      "SaveURL error",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "failed to add url",
			mockError: errors.New("unexpected error"),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlRepo := mocks.NewUrlRepository(t)

			if tc.respError == "" || tc.mockError != nil {
				urlRepo.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(tc.mockError).
					Once()
			}

			handler := New(urlRepo, nil, nil).urlSave(slogdiscard.NewDiscardLogger())

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			body := rr.Body.String()

			var resp aliasResponse

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	tests := []struct {
		name          string
		alias         string
		expectedAlias string
		respError     string
		mockError     error
	}{
		{
			name:          "Success",
			alias:         "test_alias",
			expectedAlias: "test_alias",
		},
		{
			name:          "DeleteURL error",
			alias:         "test_alias",
			expectedAlias: "",
			respError:     "url not found",
			mockError:     repository.ErrURLNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlRepo := mocks.NewUrlRepository(t)

			if tc.respError == "" || tc.mockError != nil {
				urlRepo.On("DeleteURL", tc.alias).
					Return(tc.mockError).
					Once()
			}

			r := chi.NewRouter()
			r.Delete("/{alias}", New(urlRepo, nil, nil).urlDelete(slogdiscard.NewDiscardLogger()))

			ts := httptest.NewServer(r)
			defer ts.Close()

			body, err := api.DeleteUrl(ts.URL + "/" + tc.alias)

			var aliasResp aliasResponse
			err = json.Unmarshal(body, &aliasResp)
			require.NoError(t, err)

			require.Equal(t, tc.respError, aliasResp.Error)
			require.Equal(t, tc.expectedAlias, aliasResp.Alias)
		})

	}
}

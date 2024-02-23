package redirect

import (
	"net/http/httptest"
	"testing"

	"github.com/4aykovski/learning/golang/rest/internal/http-server/handlers/url/redirect/mocks"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api"
	"github.com/4aykovski/learning/golang/rest/internal/lib/logger/handlers/slogdiscard"
	"github.com/go-chi/chi/v5"
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
			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetterMock.On("GetURL", tc.alias).
					Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToUrl, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			require.Equal(t, tc.url, redirectedToUrl)
		})
	}
}

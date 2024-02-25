package tests

import (
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/4aykovski/learning/golang/rest/internal/http-server/v1/handler"
	"github.com/4aykovski/learning/golang/rest/internal/lib/api"
	"github.com/4aykovski/learning/golang/rest/internal/lib/random"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8080"
)

func TestUrlShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/url/save").
		WithJSON(handler.UrlSaveInput{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth("myuser", "mypassword").
		Expect().Status(http.StatusOK).JSON().Object().ContainsKey("alias")
}

func TestURLShortener_SaveRedirectDelete(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "valid url",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "invalid url",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "empty alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// save
			res := e.POST("/url/save").
				WithJSON(handler.UrlSaveInput{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("myuser", "mypassword").
				Expect().Status(http.StatusOK).JSON().Object()

			if tc.error != "" {
				res.NotContainsKey("alias")

				res.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				res.Value("alias").String().IsEqual(tc.alias)
			} else {
				res.Value("alias").String().NotEmpty()

				alias = res.Value("alias").String().Raw()
			}

			// redirect

			testRedirect(t, alias, tc.url)

			// remove

			reqDel := e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("myuser", "mypassword").
				Expect().Status(http.StatusOK).JSON().Object()

			reqDel.Value("status").String().IsEqual("OK")

			// redirect again

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path.Join("url", alias),
	}

	redirectedToUrl, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToUrl)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}

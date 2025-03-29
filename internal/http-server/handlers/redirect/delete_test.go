package redirect_test

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/http-server/handlers/url"
	custom_mocks "url-shortener/internal/lib/custom-mocks"
	"url-shortener/internal/storage"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		respError string
		mockError error
	}{
		{
			name:      "Not existing alias",
			alias:     "randomAlias",
			mockError: storage.ErrUrlNotFound,
			respError: storage.ErrUrlNotFound.Error(),
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			urlDeleterMock.On("DeleteURL", tc.alias).
				Return(tc.mockError).
				Once()
			logger := slog.New(custom_mocks.NewMockLogger())
			handler := redirect.DeleteHandler(logger, urlDeleterMock)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, "/{alias}", nil)

			reqCtx := chi.NewRouteContext()
			reqCtx.URLParams.Add("alias", tc.alias)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, reqCtx))
			handler.ServeHTTP(w, r)

			require.Equal(t, w.Code, http.StatusOK)

			body := w.Body.String()

			var resp url.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}

package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=URLGetter
type URLGetter interface {
	GetURL(alias string) (string, error)
}

func GetHandler(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)
		alias := chi.URLParam(r, "alias")
		url, err := urlGetter.GetURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Error("url not found", "alias", alias, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", "alias", alias, "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}
		http.Redirect(w, r, url, http.StatusSeeOther)
	}
}

package redirect

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=URLDeleter
type URLDeleter interface {
	DeleteURL(alias string) error
}

func DeleteHandler(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)
		alias := chi.URLParam(r, "alias")
		fmt.Println("requested", r.RequestURI)
		err := urlDeleter.DeleteURL(alias)
		if errors.Is(err, storage.ErrUrlNotFound) {
			log.Error("url not found", "alias", alias, "err", err)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to get url", "alias", alias, "err", err)
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}
		log.Info("url deleted", "alias", alias)
		render.JSON(w, r, resp.OK())
	}
}

package url

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=URLSaver
type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)
		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", "err", err)
			render.JSON(w, r, resp.Error("failed to decode request body"))
			return
		}
		log.Debug("request body decoded", "body", req)
		if err := validator.New().Struct(req); err != nil {
			log.Error("failed to validate request", "err", err)
			render.JSON(w, r, resp.Error("failed to validate request"))
			return
		}

		id, alias, err := trySaveAlias(req, urlSaver)
		if errors.Is(err, storage.ErrUrlExists) {
			log.Info("url already exists", "url", req.URL)
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to save url", "err", err)
			render.JSON(w, r, resp.Error("failed to save url"))
			return
		}
		log.Info("url saved", "id", id)
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}
}

func trySaveAlias(req Request, saver URLSaver) (int64, string, error) {
	aliasProvided := true
	for {
		if req.Alias == "" {
			aliasProvided = false
			req.Alias = random.NewRandomString(resp.AliasFixedLength)
		}
		id, err := saver.SaveURL(req.URL, req.Alias)
		if errors.Is(err, storage.ErrUrlExists) && !aliasProvided {
			req.Alias = ""
			continue
		}
		return id, req.Alias, err
	}
}

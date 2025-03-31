package url

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	custom_validators "url-shortener/internal/lib/custom-validators"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty" validate:"isValidAlias"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.3 --name=URLSaver
type URLSaver interface {
	SaveURL(models.UrlShortener) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log = log.With(
			"request_id", middleware.GetReqID(r.Context()),
		)
		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error("failed to decode request body"))
			return
		}
		log.Debug("request body decoded", "body", req)

		if err := validateRequest(r, log); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		urlShortener, err := trySaveAlias(req, urlSaver)
		if errors.Is(err, storage.ErrUrlExists) {
			log.Info("url already exists", "url", req.URL)
			w.WriteHeader(http.StatusConflict)
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to save url", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error("failed to save url"))
			return
		}
		log.Info("url saved", "id", urlShortener.Id)
		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    urlShortener.Alias,
		})
	}
}

func trySaveAlias(req Request, saver URLSaver) (models.UrlShortener, error) {
	aliasProvided := true
	var urlShortener models.UrlShortener
	for {
		if req.Alias == "" {
			aliasProvided = false
			req.Alias = random.NewRandomString(resp.AliasFixedLength)
		}
		urlShortener = models.UrlShortener{
			Url:   req.URL,
			Alias: req.Alias,
		}
		id, err := saver.SaveURL(urlShortener)
		if errors.Is(err, storage.ErrUrlExists) && !aliasProvided {
			req.Alias = ""
			continue
		}
		urlShortener.Id = id
		return urlShortener, err
	}
}

func validateRequest(req *http.Request, log *slog.Logger) error {
	validate := validator.New()
	err := validate.RegisterValidation("isValidAlias", custom_validators.AliasValidation)
	if err != nil {
		log.Error("failed to register validation", "err", err)
		return errors.New("internal server error")
	}
	if err = validate.Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		err = custom_validators.ValidationError(validateErr)
		log.Error("failed to validate request", "err", err.Error())
		return err
	}
	return nil
}

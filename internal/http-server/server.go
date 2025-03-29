package http_server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url"
	middleware2 "url-shortener/internal/http-server/middleware"
)

type URLRepo interface {
	GetURL(alias string) (string, error)
	SaveURL(urlToSave string, alias string) (int64, error)
	DeleteURL(alias string) error
}

type server struct {
	router *chi.Mux
}

func New(logger *slog.Logger, repo URLRepo) *server {
	srv := &server{
		router: chi.NewRouter(),
	}
	srv.initRoutes(logger, repo)
	return srv
}

func (s *server) initRoutes(logger *slog.Logger, repo URLRepo) {

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware2.New(logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)

	s.router.Post("/url", url.New(logger, repo))
	s.router.Route("/{alias}", func(r chi.Router) {
		r.Get("/", redirect.GetHandler(logger, repo))
		r.Delete("/", redirect.DeleteHandler(logger, repo))
	})
}

func (s *server) Run(cfg *config.Config) error {
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           s.router,
		ReadHeaderTimeout: cfg.Timeout,
		WriteTimeout:      cfg.Timeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	return srv.ListenAndServe()
}

package http_server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/auth"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/url"
	middleware2 "url-shortener/internal/http-server/middleware"
	jwt_helper "url-shortener/internal/lib/jwt-helper"
	"url-shortener/internal/models"
)

type URLRepo interface {
	GetURL(alias string) (string, error)
	SaveURL(models.UrlShortener) (int64, error)
	DeleteURL(alias string) error
	SaveUser(models.User) (int64, error)
	GetUserByEmail(string) (*models.User, error)
}

type server struct {
	router *chi.Mux
	cfg    *config.Config
}

func New(logger *slog.Logger, cfg *config.Config, repo URLRepo) *server {
	srv := &server{
		router: chi.NewRouter(),
		cfg:    cfg,
	}
	srv.initRoutes(logger, repo)
	jwt_helper.InitJwtHelper(cfg)
	return srv
}

func (s *server) initRoutes(logger *slog.Logger, repo URLRepo) {

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware2.NewLoggerMW(logger))
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.URLFormat)

	s.router.Route("/", func(r chi.Router) {
		r.Use(middleware2.NewAuthMW(logger))
		r.Post("/url", url.New(logger, repo))
		r.Delete("/{alias}", redirect.DeleteHandler(logger, repo))
	})
	s.router.Get("/{alias}", redirect.GetHandler(logger, repo))
	s.router.Post("/register", auth.RegisterHandler(logger, repo))
	s.router.Post("/login", auth.LoginHandler(logger, repo))
}

func (s *server) Run() error {
	srv := &http.Server{
		Addr:              s.cfg.Addr,
		Handler:           s.router,
		ReadHeaderTimeout: s.cfg.Timeout,
		WriteTimeout:      s.cfg.Timeout,
		IdleTimeout:       s.cfg.IdleTimeout,
	}
	return srv.ListenAndServe()
}

package middleware

import (
	"context"
	"errors"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strings"
	resp "url-shortener/internal/lib/api/response"
	jwthelper "url-shortener/internal/lib/jwt-helper"
)

func NewAuthMW(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			authHeaderSplitted := strings.SplitN(authHeader, " ", 2)
			if len(authHeaderSplitted) != 2 && authHeaderSplitted[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, resp.Error("unauthorized"))
				return
			}
			rawToken := authHeaderSplitted[1]
			validToken, err := jwthelper.ValidateToken(rawToken)
			if err != nil {
				log.Error("error validating token", "err", err)
				if errors.Is(err, jwthelper.ErrInvalidToken) {
					w.WriteHeader(http.StatusUnauthorized)
					render.JSON(w, r, resp.Error("unauthorized"))
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("internal server error"))
				return
			}
			rr := context.WithValue(r.Context(), "token", validToken)
			next.ServeHTTP(w, r.WithContext(rr))
		}
		return http.HandlerFunc(fn)
	}
}

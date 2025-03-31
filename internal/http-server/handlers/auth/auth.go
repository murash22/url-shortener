package auth

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	custom_validators "url-shortener/internal/lib/custom-validators"
	jwt_helper "url-shortener/internal/lib/jwt-helper"
	"url-shortener/internal/models"
	"url-shortener/internal/storage"
)

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	Token string `json:"token"`
}

type UserRepo interface {
	SaveUser(user models.User) (int64, error)
	GetUserByEmail(email string) (*models.User, error)
}

func RegisterHandler(log *slog.Logger, repo UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := middleware.GetReqID(r.Context())
		log = log.With("request_id", reqId)

		req, err := validateRequest(r, log)
		if req == nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.MinCost)
		if err != nil {
			log.Error("failed to hash password", "err", err)
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}
		user := models.User{
			Email:    req.Email,
			Password: hashedPassword,
		}
		uid, err := repo.SaveUser(user)
		if err != nil {
			log.Error("error while saving user", "err", err)
			if errors.Is(err, storage.ErrUserExists) {
				render.JSON(w, r, resp.Error("user already exists"))
				return
			}
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}
		log.Info("user registered successfully", "user", user)
		user.Id = uid
		token, err := jwt_helper.NewToken(user)
		if err != nil {
			log.Error("failed to generate token", "err", err)
			render.JSON(w, r, resp.Error("internal server error"))
		}
		render.JSON(w, r, Response{Token: token})
	}
}

func LoginHandler(log *slog.Logger, repo UserRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqId := middleware.GetReqID(r.Context())
		log = log.With("request_id", reqId)

		req, err := validateRequest(r, log)
		if req == nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		user, err := repo.GetUserByEmail(req.Email)
		if err != nil {
			log.Error("error while getting user by email", "err", err)
			if errors.Is(err, storage.ErrUserNotFound) {
				render.JSON(w, r, resp.Error("bad credentials"))
				return
			}
			render.JSON(w, r, resp.Error("internal server error"))
			return
		}
		if err = bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password)); err != nil {
			log.Info("bad credentials", "err", err)
			render.JSON(w, r, resp.Error("bad credentials"))
			return
		}

		token, err := jwt_helper.NewToken(*user)
		if err != nil {
			log.Error("failed to generate token", "err", err)
			render.JSON(w, r, resp.Error("internal server error"))
		}
		log.Info("user login successfully", "user", user)
		render.JSON(w, r, Response{Token: token})
	}
}

func validateRequest(r *http.Request, log *slog.Logger) (*Request, error) {
	var req Request
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		log.Error("failed to decode request body", "err", err)
		return nil, errors.New("bad request")
	}
	log.Debug("request body decoded", "body", req)
	if err := validator.New().Struct(req); err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)
		err = custom_validators.ValidationError(validateErr)
		log.Error("failed to validate request", "err", err.Error())
		return nil, err
	}
	return &req, nil
}

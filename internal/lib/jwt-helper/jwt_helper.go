package jwt_helper

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
	"url-shortener/internal/config"
	"url-shortener/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

var secret string

func InitJwtHelper(cfg *config.Config) {
	secret = cfg.JwtSecret
}

func NewToken(user models.User) (string, error) {
	//token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	//	"uid":   user.Id,
	//	"email": user.Email,
	//	"exp":   time.Now().Add(time.Hour * 3).Unix(),
	//})
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &UserClaims{
		Email: user.Email,
		Id:    user.Id,
		Exp:   time.Now().Add(time.Hour * 3).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}

func ValidateToken(token string) (*UserClaims, error) {
	parser := jwt.NewParser()
	claims, err := parser.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, ErrInvalidToken
	}
	return claims.Claims.(*UserClaims), nil
}

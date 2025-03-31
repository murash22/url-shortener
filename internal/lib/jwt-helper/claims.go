package jwt_helper

import (
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type UserClaims struct {
	Id    int64  `json:"uid"`
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
}

func (c *UserClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(c.Exp, 0)), nil
}

func (c *UserClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *UserClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (c *UserClaims) GetIssuer() (string, error) {
	return "url-shortener", nil
}

func (c *UserClaims) GetSubject() (string, error) {
	uid := strconv.FormatInt(c.Id, 10)
	return uid, nil
}

func (c *UserClaims) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

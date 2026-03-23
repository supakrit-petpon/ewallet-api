package jwt

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"time"

	"github.com/golang-jwt/jwt"
)

type TokenProvider interface{
	GenerateToken(userId uint) (string, error)
}

func NewTokenProvider(SecretKey string) TokenProvider {
    return &JWTHandler{SecretKey: SecretKey}
}

type JWTHandler struct{
	SecretKey string
}

func (h *JWTHandler) GenerateToken(userId uint) (string, error){
	if h.SecretKey == ""{
		return "", errors.New("secret key is missing")
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(h.SecretKey))
	if err != nil{
		return "", domain.ErrInternalServerError
	}

	return t, nil
}
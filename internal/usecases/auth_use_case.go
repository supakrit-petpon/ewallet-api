package usecases

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/infrastructure/jwt"
	"piano/e-wallet/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)
type AuthUseCase interface{
	Login(email string, password string) (string, error)
}

type AuthService struct {
    repo          domain.UserRepository
    tokenProvider jwt.TokenProvider
	logger 		  logger.Logger
}

func NewAuthService(r domain.UserRepository, tp jwt.TokenProvider, logger logger.Logger) AuthUseCase {
    return &AuthService{repo: r, tokenProvider: tp, logger: logger}
}

func (u *AuthService) Login(email string, password string) (string, error){
	//Search user by email
	existUser, err := u.repo.Find(email)
	if err != nil {
		if errors.Is(err, domain.ErrAuthUnauthorized){
			u.logger.Warn("login failed: user not found", "error", err, "email", email)
			return "", err
		}

		u.logger.Error("login failed: internal server error", err)
		return "", err
	}

	//Compare hashpassword
	if err := bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(password)); err != nil{
		u.logger.Warn("login failed: password is incorrect", "error", err)
		return "", domain.ErrAuthUnauthorized
	}
	
	//Generate token for user
	token, err := u.tokenProvider.GenerateToken(existUser.ID)
    if err != nil {
		u.logger.Error("login failed: generate token failed", err, existUser.ID)
        return "", domain.ErrInternalServerError
    }

    return token, nil
}
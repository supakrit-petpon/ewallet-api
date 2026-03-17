package usecases

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/infrastructure/jwt"

	"golang.org/x/crypto/bcrypt"
)
type AuthUseCase interface{
	Login(email string, password string) (string, error)
}

type AuthService struct {
    repo          domain.UserRepository
    tokenProvider jwt.TokenProvider
}

func NewAuthService(r domain.UserRepository, tp jwt.TokenProvider) AuthUseCase {
    return &AuthService{repo: r, tokenProvider: tp}
}

func (u *AuthService) Login(email string, password string) (string, error){
	//Search user by email
	existUser, err := u.repo.Find(email)
	if err != nil {
		return "", err
	}

	//Compare hashpassword
	if err := bcrypt.CompareHashAndPassword([]byte(existUser.Password), []byte(password)); err != nil{
		return "", domain.ErrInvalidCredentials
	}
	
	//Generate token for user
	token, err := u.tokenProvider.GenerateToken(existUser.ID)
    if err != nil {
        return "", err
    }

    return token, nil
}
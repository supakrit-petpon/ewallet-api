package usecases

import (
	"piano/e-wallet/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

//Token
type mockTokenProvider struct{
	generateTokenFunc func(userId uint) (string, error)
}

func (m *mockTokenProvider) GenerateToken(userId uint) (string, error){
	return m.generateTokenFunc(userId)
}


func TestLogin(t *testing.T) {
	t.Run("find user success", func(t *testing.T) {
	userRepo := &mockUserRepo{
		findUserFunc: func(email string) (*domain.User, error) {
			return &domain.User{ID: 1, Email: "piano@example.com", Password: "$2a$10$aS5pRjQ6Fyx9lzdRvy8VZ.pj1Lnp23w48QCROtMPZxTx6.UUMfyc2"}, nil
		},
	}
	tokenProvider := &mockTokenProvider{
		generateTokenFunc: func(userId uint) (string, error) {
			return "valid_token", nil
		},
	}
	service := NewAuthService(userRepo, tokenProvider)
	_, err := service.Login("piano@example.com", "password")
	assert.NoError(t, err)
	})

	t.Run("Invalid email", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{
		findUserFunc: func(email string) (*domain.User, error) {
			return nil, domain.ErrInvalidCredentials
		},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				called = true
				return "valid_token", nil
			},
		}
		service := NewAuthService(userRepo, tokenProvider)
		_, err := service.Login("piano@example.com", "password")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email or password")
		assert.False(t, called, "should return error immediately and not call Generate token")
	})
	t.Run("Invalid password", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{
		findUserFunc: func(email string) (*domain.User, error) {
			return nil, domain.ErrInvalidCredentials
			},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				called = true
				return "valid_token", nil
			},
		}
		service := NewAuthService(userRepo, tokenProvider)
		_, err := service.Login("piano@example.com", "wrong_password")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email or password")
		assert.False(t, called, "should return error immediately and not call Generate token")
	})
	t.Run("Generate token error", func(t *testing.T) {
		userRepo := &mockUserRepo{
		findUserFunc: func(email string) (*domain.User, error) {
			return &domain.User{ID: 1, Email: "piano@example.com", Password: "$2a$10$aS5pRjQ6Fyx9lzdRvy8VZ.pj1Lnp23w48QCROtMPZxTx6.UUMfyc2"}, nil
		},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				return "", domain.ErrInternalServerError
			},
		}
		service := NewAuthService(userRepo, tokenProvider)
		_, err := service.Login("piano@example.com", "password")
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
		
	})

}

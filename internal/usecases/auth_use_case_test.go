package usecases

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
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
	testLog := logger.NewTestLogger(t)

	t.Run("login success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			findFunc: func(email string) (*domain.User, error) {
				return &domain.User{ID: 1, Email: "piano@example.com", Password: "$2a$10$aS5pRjQ6Fyx9lzdRvy8VZ.pj1Lnp23w48QCROtMPZxTx6.UUMfyc2"}, nil
			},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				return "valid_token", nil
			},
		}
		service := NewAuthService(userRepo, tokenProvider, testLog)
		token, err := service.Login("piano@example.com", "password")

		assert.NoError(t, err)
		assert.NotNil(t, token)
	})
	t.Run("Invalid email", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{
			findFunc: func(email string) (*domain.User, error) {
				return nil, domain.ErrNotFoundUser
			},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				called = true
				return "valid_token", nil
			},
		}
		service := NewAuthService(userRepo, tokenProvider, testLog)
		_, err := service.Login("piano@example.com", "password")
		
		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundUser, err)
		assert.False(t, called, "should return error immediately and not call Generate token")
	})
	t.Run("Invalid password", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{
			findFunc: func(email string) (*domain.User, error) {
				return nil, domain.ErrNotFoundUser
				},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				called = true
				return "valid_token", nil
			},
		}
		service := NewAuthService(userRepo, tokenProvider, testLog)
		_, err := service.Login("piano@example.com", "wrong_password")
		
		assert.Error(t, err)
		assert.Error(t, domain.ErrNotFoundUser, err)
		assert.False(t, called, "should return error immediately and not call Generate token")
	})
	t.Run("Generate token error", func(t *testing.T) {
		userRepo := &mockUserRepo{
		findFunc: func(email string) (*domain.User, error) {
			return &domain.User{ID: 1, Email: "piano@example.com", Password: "$2a$10$aS5pRjQ6Fyx9lzdRvy8VZ.pj1Lnp23w48QCROtMPZxTx6.UUMfyc2"}, nil
		},
		}
		tokenProvider := &mockTokenProvider{
			generateTokenFunc: func(userId uint) (string, error) {
				return "", domain.ErrInternalServerError
			},
		}
		service := NewAuthService(userRepo, tokenProvider, testLog)
		_, err := service.Login("piano@example.com", "password")
		
		assert.Error(t, err)
		assert.Error(t, domain.ErrInternalServerError, err)
		
	})

}

package usecases

import (
	"piano/e-wallet/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWalletRepo struct {
	createFunc func(wallet domain.Wallet) error
	getFunc func(id uint) (int64, error)
}

func (m *mockWalletRepo) Get(id uint) (int64, error) {
	return m.getFunc(id)
}

func (m *mockWalletRepo) Create(wallet domain.Wallet) error {
	return m.createFunc(wallet)
}


func TestGetBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (int64, error) {
				return 100000, nil
			},
		}
		service := NewWalletService(repo)

		balance, err := service.Balance(1)

		assert.NoError(t, err)
		assert.Equal(t, "1000.00 THB", balance)
	})

	t.Run("invalid user id", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (int64, error) {
				return 0, domain.ErrUserRecordNotFound
			},
		}
		service := NewWalletService(repo)
		_, err := service.Balance(999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid user id")
	})

	t.Run("internal server error", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (int64, error) {
				return 0, domain.ErrInternalServerError
			},
		}
		service := NewWalletService(repo)
		_, err := service.Balance(1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
	})
}

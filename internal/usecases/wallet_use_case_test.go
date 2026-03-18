package usecases

import (
	"piano/e-wallet/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWalletRepo struct {
	createWalletFunc func(wallet domain.Wallet) error
	getBalanceFunc func(id uint) (int64, error)
}

func (m *mockWalletRepo) GetBalance(id uint) (int64, error) {
	return m.getBalanceFunc(id)
}

func (m *mockWalletRepo) CreateWallet(wallet domain.Wallet) error {
	return m.createWalletFunc(wallet)
}


func TestGetBalance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockWalletRepo{
			getBalanceFunc: func(id uint) (int64, error) {
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
			getBalanceFunc: func(id uint) (int64, error) {
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
			getBalanceFunc: func(id uint) (int64, error) {
				return 0, domain.ErrInternalServerError
			},
		}
		service := NewWalletService(repo)
		_, err := service.Balance(1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Internal server error")
	})
}

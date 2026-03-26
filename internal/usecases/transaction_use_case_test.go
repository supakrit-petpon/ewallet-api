package usecases

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransactionRepo struct{
	createTxFunc func(tx domain.Transaction) error
	updateTxFunc func(id uint, status string) (*domain.Transaction, error)
	getTxFunc func(refId string) (*domain.Transaction, error)
	getAllTxFunc func(userId uint) ([]domain.Transaction, error)
}

func (m *mockTransactionRepo) Create(tx *domain.Transaction) error{
	return m.createTxFunc(*tx)
}
func (m *mockTransactionRepo) Update(id uint, status string) (*domain.Transaction, error){
	return m.updateTxFunc(id, status)
}
func (m *mockTransactionRepo) Get(refId string) (*domain.Transaction, error){
	return m.getTxFunc(refId)
}
func (m *mockTransactionRepo) GetAll(userId uint) ([]domain.Transaction, error){
	return m.getAllTxFunc(userId)
}


func TestGetTransaction(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	t.Run("success", func(t *testing.T) {
		refID := "refID"
		txRepo := &mockTransactionRepo{
			getTxFunc: func(refId string) (*domain.Transaction, error) {
				return &domain.Transaction{ReferenceID: refId}, nil
			},
		}
		walletRepo := &mockWalletRepo{}
		
		service := NewTransactionService(txRepo, walletRepo, testLog)
		tx, err := service.GetTransaction(refID)

		assert.NoError(t, err)
		assert.Equal(t, refID, tx.ReferenceID)
	})
	t.Run("failure: transaction not found", func(t *testing.T) {
		refID := "refID"
		txRepo := &mockTransactionRepo{
			getTxFunc: func(refId string) (*domain.Transaction, error) {
				return nil, domain.ErrNotFoundTransaction
			},
		}
		walletRepo := &mockWalletRepo{}
		service := NewTransactionService(txRepo, walletRepo, testLog)
		tx, err := service.GetTransaction(refID)

		assert.Error(t, err)
		assert.Nil(t, tx)
	})
	t.Run("failure: internal server error", func(t *testing.T) {
		refID := "refID"
		txRepo := &mockTransactionRepo{
			getTxFunc: func(refId string) (*domain.Transaction, error) {
				return nil, domain.ErrInternalServerError
			},
		}
		walletRepo := &mockWalletRepo{}

		service := NewTransactionService(txRepo, walletRepo, testLog)
		tx, err := service.GetTransaction(refID)

		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}

func TestGettAllTransaction(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	t.Run("success", func(t *testing.T) {
		userId := uint(1)
		walletRepo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return &domain.Wallet{}, nil
			},
		}
		txRepo := &mockTransactionRepo{
			getAllTxFunc: func(userId uint) ([]domain.Transaction, error) {
				transactions := []domain.Transaction{
					{TransactionType: "TOPUP"},
					{TransactionType: "WITHDRAW"},
				}
				return transactions, nil
			},
		}

		service := NewTransactionService(txRepo, walletRepo, testLog)
		tx, err := service.GetAllTransaction(userId)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(tx))
	})
	t.Run("failure: get wallet id fail", func(t *testing.T) {
		userId := uint(1)
		walletRepo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return nil, domain.ErrNotFoundWallet
			},
		}
		txRepo := &mockTransactionRepo{}

		service := NewTransactionService(txRepo, walletRepo, testLog)
		_, err := service.GetAllTransaction(userId)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
	})
	t.Run("failure: internal server error", func(t *testing.T) {
		userId := uint(1)
		walletRepo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return &domain.Wallet{}, nil
			},
		}
		txRepo := &mockTransactionRepo{
			getAllTxFunc: func(userId uint) ([]domain.Transaction, error) {
				return nil, domain.ErrInternalServerError
			},
		}

		service := NewTransactionService(txRepo, walletRepo, testLog)
		_, err := service.GetAllTransaction(userId)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
	})
}
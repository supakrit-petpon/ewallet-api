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


func TestGetTransaction(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	t.Run("success", func(t *testing.T) {
		refID := "refID"
		txRepo := &mockTransactionRepo{
			getTxFunc: func(refId string) (*domain.Transaction, error) {
				return &domain.Transaction{ReferenceID: refId}, nil
			},
		}
		
		service := NewTransactionService(txRepo, testLog)
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
		
		service := NewTransactionService(txRepo, testLog)
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
		
		service := NewTransactionService(txRepo, testLog)
		tx, err := service.GetTransaction(refID)

		assert.Error(t, err)
		assert.Nil(t, tx)
	})
}
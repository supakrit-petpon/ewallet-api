package usecases

import "piano/e-wallet/internal/domain"

type mockTransactionRepo struct{
	createTxFunc func(tx domain.Transaction) error
	updateTxFunc func(id uint, status string) (*domain.Transaction, error)
}

func (m *mockTransactionRepo) Create(tx *domain.Transaction) error{
	return m.createTxFunc(*tx)
}
func (m *mockTransactionRepo) Update(id uint, status string) (*domain.Transaction, error){
	return m.updateTxFunc(id, status)
}

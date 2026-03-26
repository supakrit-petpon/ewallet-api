package usecases

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
)

type TransactionUseCase interface{
	GetTransaction(refId string) (*domain.Transaction, error)
}

type TransactionService struct {
	repo domain.TransactionRepository
	logger logger.Logger
}

func NewTransactionService(repo domain.TransactionRepository, logger logger.Logger) TransactionUseCase{
	return &TransactionService{repo: repo, logger: logger}
}

func (s *TransactionService) GetTransaction(refId string) (*domain.Transaction, error){
	transaction, err := s.repo.Get(refId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundTransaction){
			s.logger.Warn("Get transaction failed: transaction record not found", err)
			return nil, err
		}
		
		s.logger.Error("Get transaction failed: database connection lost during create user", err)
		return nil, err
	}

	return transaction, nil
}
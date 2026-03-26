package usecases

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
)

type TransactionUseCase interface{
	GetTransaction(refId string) (*domain.Transaction, error)
	GetAllTransaction(userId uint) ([]domain.Transaction, error)
}

type TransactionService struct {
	txRepo domain.TransactionRepository
	walletRepo domain.WalletRepository
	logger logger.Logger
}

func NewTransactionService(txRepo domain.TransactionRepository,walletRepo domain.WalletRepository, logger logger.Logger) TransactionUseCase{
	return &TransactionService{txRepo: txRepo, walletRepo: walletRepo, logger: logger}
}

func (s *TransactionService) GetTransaction(refId string) (*domain.Transaction, error){
	transaction, err := s.txRepo.Get(refId)
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

func (s *TransactionService) GetAllTransaction(userId uint) ([]domain.Transaction, error){
	wallet, err := s.walletRepo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("Get all transaction: wallet record not found", err, "user_id", userId)
			return nil, err
		}

		s.logger.Error("Get all transaction: database connection lost during get wallet", err)
		return nil, err
	}

	transactions, err := s.txRepo.GetAll(wallet.ID)
	if err != nil {
		s.logger.Error("Get all transaction failed: database connection lost during get all transaction", err)
		return nil, err
	}

	return transactions, nil
}
package usecases

import (
	"errors"
	"fmt"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"time"
)

type WalletUseCase interface{
	Balance(userId uint) (string, error)
	TopUp(userId uint, amount float64) (*domain.Transaction, float64, error)
}

type WalletService struct{
	repo domain.WalletRepository
	txRepo domain.TransactionRepository
	logger logger.Logger
}

func NewWalletService(repo domain.WalletRepository, txRepo domain.TransactionRepository, logger logger.Logger) WalletUseCase{
	return &WalletService{repo: repo, txRepo: txRepo, logger: logger}
}

func (s *WalletService) Balance(userId uint) (string, error){
	wallet, err := s.repo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("balance failed: wallet record not found", err, "user_id", userId)
			return "", err
		}

		s.logger.Error("balance failed: database connection lost during get wallet", err)
		return "", err
	}

	formattedBalance := fmt.Sprintf("%.2f THB", float64(wallet.Balance)/100)
	return formattedBalance, nil
}

func (s *WalletService) TopUp(userId uint, amount float64) (*domain.Transaction, float64, error) {
    refID := fmt.Sprintf("TOPUP-%d-%d", userId, time.Now().UnixNano())
	
	//1. Get Wallet Id
	wallet, err := s.repo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("balance failed: wallet record not found", err, "user_id", userId)
			return nil, 0, err
		}

		s.logger.Error("balance failed: database connection lost during get wallet", err)
		return nil, 0, err
	}
	
	//2. สร้าง transaction ด้วย status 'PENDING'
	newTx := &domain.Transaction{
				DestinationID:   &wallet.ID,
				Amount:          int(amount * 100),
				TransactionType: "TOPUP",
				Status: "PENDING",
				ReferenceID:     refID,
	}

	err = s.txRepo.Create(newTx);
	if err != nil {
		if errors.Is(err, domain.ErrConflictTransactionRefId){
			s.logger.Warn("TopUp failed: this transaction is already created", err, "component", "wallet service")
			return nil, 0, err
		}

		s.logger.Error("TopUp failed: database connection lost during create user", err)
		return nil, 0, err
	}
	
	//3. เริ่ม Database Transaction
	tran, balance, err := s.repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {

		updatedBalance, err := txWallet.IncrementBalance(userId, int64(amount*100))	
		if err != nil {
			if errors.Is(err, domain.ErrNotFoundWallet){
				s.logger.Warn("TopUp failed: wallet record not found", err)
				return nil, 0, err
			}

			s.logger.Error("TopUp failed: database connection lost during increment balance", err)
			return nil, 0, err
		}	

		//Update transaction ด้วย status 'SUCCESS'
		transaction, err := txTrans.Update(newTx.ID, "SUCCESS"); if err != nil {
			if errors.Is(err, domain.ErrNotFoundTransaction){
				s.logger.Warn("TopUp failed: transaction record not found", err)
				return nil, 0, err
			}

			s.logger.Error("TopUp failed: database connection lost during updating transaction", err)
			return nil, 0, err
		}
	
		return transaction, float64(updatedBalance),  nil
	})

	//4. ถ้า transaction พัง ให้ทำ Error Handling
	if err != nil {
		//Update transaction ด้วย status 'FAILED'
		tran, _ = s.txRepo.Update(newTx.ID, "FAILED")
		return tran, 0, err
	}

	return tran, balance , nil
}
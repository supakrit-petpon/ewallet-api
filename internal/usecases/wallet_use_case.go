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
	Withdraw(userId uint, amount float64) (*domain.Transaction, float64, error)
	Transfer(userId uint, desId uint, amount float64) (*domain.Transaction, float64, error)
	Info(userId uint) (uint, error)
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
			s.logger.Warn("topup failed: wallet record not found", err, "user_id", userId)
			return nil, 0, err
		}

		s.logger.Error("topup failed: database connection lost during get wallet", err)
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
	transaction, balance, err := s.repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {

		updatedBalance, err := txWallet.IncrementBalance(wallet.ID, int64(amount*100))	
		if err != nil {
			if errors.Is(err, domain.ErrNotFoundWallet){
				s.logger.Warn("TopUp failed: wallet record not found", err)
				return nil, 0, err
			}

			s.logger.Error("TopUp failed: database connection lost during increment balance", err)
			return nil, 0, err
		}	

		//Update transaction ด้วย status 'SUCCESS'
		updatedTransaction, err := txTrans.Update(newTx.ID, "SUCCESS"); if err != nil {
			if errors.Is(err, domain.ErrNotFoundTransaction){
				s.logger.Warn("TopUp failed: transaction record not found", err)
				return nil, 0, err
			}

			s.logger.Error("TopUp failed: database connection lost during updating transaction", err)
			return nil, 0, err
		}
	
		return updatedTransaction, float64(updatedBalance),  nil
	})

	//4. ถ้า transaction พัง ให้ทำ Error Handling
	if err != nil {
		//Update transaction ด้วย status 'FAILED'
		failTransaction, updateErr := s.txRepo.Update(newTx.ID, "FAILED")
		if updateErr != nil {
			return newTx, 0, err
		}
		return  failTransaction, 0, err
	}

	return transaction, balance , nil
}

func (s *WalletService) Withdraw(userId uint, amount float64) (*domain.Transaction, float64, error){
	refID := fmt.Sprintf("WITHDRAW-%d-%d", userId, time.Now().UnixNano())

	//1. Get Wallet Id
	wallet, err := s.repo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("Withdraw failed: wallet record not found", err, "user_id", userId)
			return nil, 0, err
		}

		s.logger.Error("Withdraw failed: database connection lost during get wallet", err)
		return nil, 0, err
	}

	newTx := &domain.Transaction{
				SourceID:   &wallet.ID,
				Amount:          int(amount * 100),
				TransactionType: "WITHDRAW",
				Status: "PENDING",
				ReferenceID:     refID,
	}

	err = s.txRepo.Create(newTx);
	if err != nil {
		if errors.Is(err, domain.ErrConflictTransactionRefId){
			s.logger.Warn("Withdraw failed: this transaction is already created", err, "component", "wallet service")
			return nil, 0, err
		}

		s.logger.Error("Withdraw failed: database connection lost during create user", err)
		return nil, 0, err
	}

	//3. เริ่ม Database Transaction
	transaction, balance, err := s.repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {

		updatedBalance, err := txWallet.DecrementBalance(wallet.ID, int64(amount*100))	
		if err != nil {
			if errors.Is(err, domain.ErrInsufficientBalance){
				s.logger.Warn("Withdraw failed: insufficient balance for this transaction", err)
				return nil, 0, err
			}

			s.logger.Error("Withdraw failed: database connection lost during increment balance", err)
			return nil, 0, err
		}	

		//Update transaction ด้วย status 'SUCCESS'
		transaction, err := txTrans.Update(newTx.ID, "SUCCESS"); if err != nil {
			if errors.Is(err, domain.ErrNotFoundTransaction){
				s.logger.Warn("Withdraw failed: transaction record not found", err)
				return nil, 0, err
			}

			s.logger.Error("Withdraw failed: database connection lost during updating transaction", err)
			return nil, 0, err
		}
	
		return transaction, float64(updatedBalance),  nil
	})

	//4. ถ้า transaction พัง ให้ทำ Error Handling
	if err != nil {
		//Update transaction ด้วย status 'FAILED'
		failTransaction, updateErr := s.txRepo.Update(newTx.ID, "FAILED")
		if updateErr != nil {
			return newTx, 0, err
		}
		return  failTransaction, 0, err
	}

	return transaction, balance , nil
}

func (s *WalletService) Transfer(userId uint, desId uint, amount float64) (*domain.Transaction, float64, error){
	refID := fmt.Sprintf("TRANSFER-%d-%d", userId, time.Now().UnixNano())

	//1. Get Wallet Id
	wallet, err := s.repo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("Transfer failed: wallet record not found", err, "user_id", userId)
			return nil, 0, err
		}

		s.logger.Error("Transfer failed: database connection lost during get wallet", err)
		return nil, 0, err
	}

	//2. Check wallet_id and des_id
	if wallet.ID == desId {
		s.logger.Warn("Transfer failed: source_id and destination_id can't be same")
		return nil, 0, domain.ErrConflictSourceDesId
	}

	//3. Create Transaction
	newTx := &domain.Transaction{
				SourceID:   &wallet.ID,
				DestinationID: &desId,
				Amount:          int(amount * 100),
				TransactionType: "TRANSFER",
				Status: 		"PENDING",
				ReferenceID:     refID,
	}
	err = s.txRepo.Create(newTx);
	if err != nil {
		if errors.Is(err, domain.ErrConflictTransactionRefId){
			s.logger.Warn("Withdraw failed: this transaction is already created", err, "component", "wallet service")
			return nil, 0, err
		}

		s.logger.Error("Withdraw failed: database connection lost during create user", err)
		return nil, 0, err
	}

	//4. เริ่ม Database Transaction
	transaction, balance, err := s.repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {

		//3.1 ลดเงินของเจ้าของบัญชี
		updatedBalance, err := txWallet.DecrementBalance(wallet.ID, int64(amount * 100))	
		if err != nil {
			if errors.Is(err, domain.ErrInsufficientBalance){
				s.logger.Warn("Transfer failed: insufficient balance for this transaction for", err)
				return nil, 0, err
			}

			s.logger.Error("Transfer failed: database connection lost during increment balance", err)
			return nil, 0, err
		}	
		
		//3.2 เพิ่มเงินบัญชีปลายทาง
		_, err = txWallet.IncrementBalance(desId, int64(amount * 100))
		if err != nil {
			if errors.Is(err, domain.ErrNotFoundWallet){
				s.logger.Warn("TopUp failed: wallet record not found", err)
				return nil, 0, err
			}

			s.logger.Error("TopUp failed: database connection lost during increment balance", err)
			return nil, 0, err
		}
 
		// Update transaction ด้วย status 'SUCCESS'
		transaction, err := txTrans.Update(newTx.ID, "SUCCESS"); if err != nil {
			if errors.Is(err, domain.ErrNotFoundTransaction){
				s.logger.Warn("Withdraw failed: transaction record not found", err)
				return nil, 0, err
			}

			s.logger.Error("Withdraw failed: database connection lost during updating transaction", err)
			return nil, 0, err
		}
	
		return transaction, float64(updatedBalance),  nil
	})

	//5. ถ้า transaction พัง ให้ทำ Error Handling
	if err != nil {
		//Update transaction ด้วย status 'FAILED'
		failTransaction, updateErr := s.txRepo.Update(newTx.ID, "FAILED")
		if updateErr != nil {
			return newTx, 0, err
		}
		return  failTransaction, 0, err
	}

	return transaction, balance, err
}

func (s *WalletService) Info(userId uint) (uint, error){
	wallet, err := s.repo.Get(userId)
	if err != nil {
		if errors.Is(err, domain.ErrNotFoundWallet){
			s.logger.Warn("Transfer failed: wallet record not found", err, "user_id", userId)
			return 0, err
		}

		s.logger.Error("Transfer failed: database connection lost during get wallet", err)
		return 0, err
	}

	return wallet.ID, nil
}
package repository

import (
	"errors"
	"fmt"
	"piano/e-wallet/internal/domain"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormWalletRepository struct{
	db *gorm.DB
}

func NewGormWalletRepository(db *gorm.DB) domain.WalletRepository{
	return &GormWalletRepository{db: db}
}

func (r *GormWalletRepository) Create(wallet domain.Wallet) error{
	result := r.db.Create(&wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) || strings.Contains(result.Error.Error(), "duplicate") {
            return domain.ErrConflictUserWallet
        }
		if errors.Is(result.Error, gorm.ErrForeignKeyViolated) || strings.Contains(result.Error.Error(), "violates foreign key constraint") {
            return domain.ErrNotFoundUser
        }

		return domain.ErrInternalServerError
	}

	return nil
}

func (r *GormWalletRepository) Get(userId uint) (*domain.Wallet, error) {
	wallet := new(domain.Wallet)

	result := r.db.Where("user_id = ?", userId).Select("id", "balance").Find(&wallet)
	
	//Internal DB error
	if result.Error != nil {
		return nil, domain.ErrInternalServerError
	}
	
	if result.RowsAffected == 0{
		return nil, domain.ErrNotFoundWallet
	}
	
	return wallet, nil
}

func (r *GormWalletRepository) IncrementBalance(walletId uint, amount int64) (int64, error){
	var wallet domain.Wallet
    
    result := r.db.Model(&wallet).
        Clauses(clause.Returning{Columns: []clause.Column{{Name: "balance"}}}).
        Where("id = ?", walletId).
        Update("balance", gorm.Expr("balance + ?", amount))

	if result.RowsAffected == 0 {
		return 0, domain.ErrNotFoundWallet
	}
    if result.Error != nil {
		fmt.Println(result.Error)
		return 0, domain.ErrInternalServerError
	}
	
	
    return int64(wallet.Balance), nil
}

func (r *GormWalletRepository) DecrementBalance(userId uint, amount int64) (int64, error){
	var wallet domain.Wallet
    
    result := r.db.Model(&wallet).
        Clauses(clause.Returning{Columns: []clause.Column{{Name: "balance"}}}).
        Where("id = ? AND balance >= ?", userId, amount).
        Update("balance", gorm.Expr("balance - ?", amount))

    if result.Error != nil {
		return 0, domain.ErrInternalServerError
	}
	if result.RowsAffected == 0 {
		return 0, domain.ErrInsufficientBalance
	}
	
    return int64(wallet.Balance), nil
}

func (r *GormWalletRepository)ExecuteTransaction(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error){
	var transaction *domain.Transaction
	var balance float64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		walletRepoTx := NewGormWalletRepository(tx)
		transRepoTx := NewGormTransactionRepository(tx)

		updatedTransaction, updatedBalance, err := fn(walletRepoTx, transRepoTx)
		if err != nil {
            return err
        }

        transaction = updatedTransaction 
		balance = updatedBalance
        return nil
	})
	if err != nil {
		return nil, 0, err
	}
	
	return transaction, balance, nil
}
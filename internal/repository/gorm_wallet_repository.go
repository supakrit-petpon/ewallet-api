package repository

import (
	"errors"
	"fmt"
	"piano/e-wallet/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type GormWalletRepository struct{
	db *gorm.DB
}

func NewGormWalletRepository(db *gorm.DB) domain.WalletRepository{
	return &GormWalletRepository{db: db}
}

func (r *GormWalletRepository) CreateWallet(wallet domain.Wallet) error{
	result := r.db.Create(&wallet)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) || strings.Contains(result.Error.Error(), "duplicate") {
            return fmt.Errorf("user already has wallet: %w", gorm.ErrDuplicatedKey)
        }
		if errors.Is(result.Error, gorm.ErrForeignKeyViolated) || strings.Contains(result.Error.Error(), "violates foreign key constraint") {
            return fmt.Errorf("user not found: %w", gorm.ErrForeignKeyViolated)
        }
	}

	return nil
}

func (r *GormWalletRepository) GetBalance(userId uint) (int64, error) {
	wallet := new(domain.Wallet)

	result := r.db.Where("user_id = ?", userId).Select("balance").Find(&wallet)
	
	//Internal DB error
	if result.Error != nil {
		return 0, domain.ErrInternalServerError
	}
	
	if result.RowsAffected == 0{
		return 0, domain.ErrUserRecordNotFound
	}
	
	return wallet.Balance, nil
}
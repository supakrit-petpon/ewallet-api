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
package domain

import (
	"gorm.io/gorm"
)

type Wallet struct{
	gorm.Model
	Balance int64 `json:"balance"`
	Currency string `gorm:"size:3;not null;default:'THB'"`
	
	UserID uint `gorm:"uniqueIndex;not null"`
}

type WalletRepository interface{
	Create(wallet Wallet) error
	Get(userId uint) (int64, error)
}

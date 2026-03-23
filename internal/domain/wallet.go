package domain

import (
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	Balance  int64  `json:"balance"`
	Currency string `gorm:"size:3;not null;default:'THB'"`

	UserID               uint          `gorm:"uniqueIndex;not null"`
	SentTransactions     []Transaction `gorm:"foreignKey:SourceID" json:"sent_transactions,omitempty"`
	ReceivedTransactions []Transaction `gorm:"foreignKey:DestinationID" json:"received_transactions,omitempty"`
}

type WalletRepository interface {
	Create(wallet Wallet) error
	Get(userId uint) (*Wallet, error)
	IncrementBalance(userId uint, amount int64) (int64, error)

	//DB transaction to update wallet & create transaction
	ExecuteTransaction(fn func(txWallet WalletRepository, txTrans TransactionRepository) (*Transaction, float64, error)) (*Transaction, float64, error)
}
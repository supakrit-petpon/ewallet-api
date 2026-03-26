package domain

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	SourceID        *uint             `gorm:"index" json:"source_id"`      // NULL หากเป็น TOPUP
	DestinationID   *uint             `gorm:"index" json:"destination_id"` // NULL หากเป็น WITHDRAW
	Amount          int               `gorm:"not null" json:"amount"`
	TransactionType string   		  `gorm:"not null" json:"transaction_type"`
	Status          string 			  `gorm:"default:'PENDING';not null" json:"status"`
	ReferenceID     string            `gorm:"unique;not null" json:"reference_id"`
	
	// Relation กลับไปที่ Wallet
	SourceWallet      *Wallet `gorm:"foreignKey:SourceID;references:ID" json:"source_wallet,omitempty"`
	DestinationWallet *Wallet `gorm:"foreignKey:DestinationID;references:ID" json:"destination_wallet,omitempty"`
}

type TransactionRepository interface{
	Create(transaction *Transaction) error
	Update(id uint, status string) (*Transaction, error)
	Get(refId string) (*Transaction, error)
}
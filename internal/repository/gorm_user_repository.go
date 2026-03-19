package repository

import (
	"errors"
	"fmt"
	"piano/e-wallet/internal/domain"
	"strings"

	"gorm.io/gorm"
)

type GormUserRepository struct{
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) domain.UserRepository{
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(user domain.User) (uint, error) {
	
	result := r.db.Create(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) || strings.Contains(result.Error.Error(), "duplicate") {
            return 0, fmt.Errorf("email is already exists: %w", gorm.ErrDuplicatedKey)
        }
	}

	return user.ID, nil
}

func (r *GormUserRepository) Transaction_CreateUser_CreateWallet(fn func(domain.UserRepository, domain.WalletRepository) error) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        userRepoTx := NewGormUserRepository(tx)
        walletRepoTx := NewGormWalletRepository(tx)
        
        return fn(userRepoTx, walletRepoTx)
    })
}

func (r *GormUserRepository) Find(email string) (*domain.User, error){
	selectedUser := new(domain.User)

	result := r.db.Where(`email = ?`, email).First(selectedUser)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
	}
	
	return selectedUser, nil
}
package usecases

import (
	"piano/e-wallet/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface{
	Register(user domain.User) error
}

type UserService struct{
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) UserUseCase{
	return &UserService{repo: repo}
}

func (s *UserService) Register(user domain.User) error {

	//1. hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	//2. use transaction to create User & Wallet
	return s.repo.Transaction_CreateUser_CreateWallet(func(txUser domain.UserRepository, txWallet domain.WalletRepository) error {
		
		id, err := txUser.CreateUser(user)
		if err != nil{
			return err
		}

		wallet := &domain.Wallet{
            UserID:   id,
            Balance:  0,
            Currency: "THB",
        }

		if err := txWallet.CreateWallet(*wallet); err != nil{
			return err
		}

		return nil
	})
	
}


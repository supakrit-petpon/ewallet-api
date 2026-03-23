package usecases

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type UserUseCase interface{
	Register(user domain.User) error
}

type UserService struct{
	repo domain.UserRepository
	logger logger.Logger
}

func NewUserService(repo domain.UserRepository, logger logger.Logger) UserUseCase{
	return &UserService{repo: repo, logger: logger}
}

func (s *UserService) Register(user domain.User) error {

	//1. hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("register failed: hashed password error", err)
		return domain.ErrInternalServerError
	}
	user.Password = string(hashedPassword)

	//2. use transaction to create User & Wallet
	return s.repo.ExecuteTransaction(func(txUser domain.UserRepository, txWallet domain.WalletRepository) error {
		
		id, err := txUser.Create(user)
		if err != nil{
			if errors.Is(err, domain.ErrConflictEmail){
				s.logger.Warn("register failed: this email is already registered", err)
				return err
			}
			s.logger.Error("register failed: database connection lost during create user", err)
			return err
		}

		wallet := &domain.Wallet{
            UserID:   id,
            Balance:  0,
            Currency: "THB",
        }

		if err := txWallet.Create(*wallet); err != nil{
			if errors.Is(err, domain.ErrConflictUserWallet){
				s.logger.Warn("register failed: this user is already has wallet", err, "user_id", id)
				return err
			}
			if errors.Is(err, domain.ErrNotFoundUser){
				s.logger.Warn("register failed: user record not found", err, "user_id", id)
				return err
			}

			s.logger.Error("register failed: database connection lost during create wallet", err, "user_id", id)
			return err
		}

		return nil
	})
	
}


package usecases

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct{
	createFunc func(user domain.User) (uint, error)
	tracsaction_CreateUser_CreateWallet func(func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error
	findFunc func(email string) (*domain.User, error)
}

func (m *mockUserRepo) Create(user domain.User) (uint, error){
	return m.createFunc(user)
}

func (m *mockUserRepo) Transaction_CreateUser_CreateWallet(fn func(domain.UserRepository, domain.WalletRepository) error) error {
    return m.tracsaction_CreateUser_CreateWallet(fn)
}

func (m *mockUserRepo) Find(email string) (*domain.User, error) {
    return m.findFunc(email)
}


func TestRegister(t *testing.T){
	//success
	t.Run("success registration with transaction", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}
		service := NewUserService(userRepo)

		userRepo.tracsaction_CreateUser_CreateWallet = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 1, nil
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }

		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.NoError(t, err)
	})
	
	//error: hashing password
	t.Run("hashing password failure", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.tracsaction_CreateUser_CreateWallet = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			called = true
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			called = true
			return 0, nil
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }
		service := NewUserService(userRepo)

		veryLongPassword := string(make([]byte, 73))

		err := service.Register(domain.User{Email: "piano@example.com", Password: veryLongPassword})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password length exceeds")
		assert.False(t, called, "should return error immediately and not call Tracsaction")
	})

	//error: repo error eg: email exist
	t.Run("email is already exists", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.tracsaction_CreateUser_CreateWallet = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 0, errors.New("email is already exists")
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }
		service := NewUserService(userRepo)

		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "email is already exists")
	})
	
	t.Run("create wallet failure: user not found", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.tracsaction_CreateUser_CreateWallet = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 100, nil
		}
		walletRepo.createFunc = func(wallet domain.Wallet) error {return errors.New("user not found")}
		service := NewUserService(userRepo)
		
		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "user not found")
	})
}
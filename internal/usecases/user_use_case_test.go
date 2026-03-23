package usecases

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockUserRepo struct{
	createFunc func(user domain.User) (uint, error)
	executeTransactionFunc func(func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error
	findFunc func(email string) (*domain.User, error)
}

func (m *mockUserRepo) Create(user domain.User) (uint, error){
	return m.createFunc(user)
}

func (m *mockUserRepo) ExecuteTransaction(fn func(domain.UserRepository, domain.WalletRepository) error) error {
    return m.executeTransactionFunc(fn)
}

func (m *mockUserRepo) Find(email string) (*domain.User, error) {
    return m.findFunc(email)
}


func TestRegister(t *testing.T){
	testLog := logger.NewTestLogger(t)
	t.Run("success registration with transaction", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}
		service := NewUserService(userRepo, testLog)

		userRepo.executeTransactionFunc = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 1, nil
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }

		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.NoError(t, err)
	})
	
	t.Run("hashing password failure", func(t *testing.T) {
		called := false
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.executeTransactionFunc = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			called = true
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			called = true
			return 0, nil
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }
		service := NewUserService(userRepo, testLog)

		veryLongPassword := string(make([]byte, 73))

		err := service.Register(domain.User{Email: "piano@example.com", Password: veryLongPassword})
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.False(t, called, "should return error immediately and not call Tracsaction")
	})

	t.Run("email is already exists", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.executeTransactionFunc = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 0, domain.ErrConflictEmail
		}
		walletRepo.createFunc = func(w domain.Wallet) error { return nil }
		service := NewUserService(userRepo, testLog)

		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.Error(t, err)
		assert.Equal(t, domain.ErrConflictEmail, err)
	})
	
	t.Run("create wallet failure: user not found", func(t *testing.T) {
		userRepo := &mockUserRepo{}
		walletRepo := &mockWalletRepo{}

		userRepo.executeTransactionFunc = func(fn func(txUser domain.UserRepository, txWallet domain.WalletRepository) error) error {
			return fn(userRepo, walletRepo)
		}
		userRepo.createFunc = func(user domain.User) (uint, error) {
			return 100, nil
		}
		walletRepo.createFunc = func(wallet domain.Wallet) error {return domain.ErrNotFoundUser}
		service := NewUserService(userRepo, testLog)
		
		err := service.Register(domain.User{Email: "piano@example.com", Password: "password"})
		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundUser, err)
	})
}
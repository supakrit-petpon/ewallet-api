package usecases

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockWalletRepo struct {
	createFunc func(wallet domain.Wallet) error
	getFunc func(userId uint) (*domain.Wallet, error)
	incrementBalanceFunc func(userId uint, amount int64) (int64, error)
	decrementBalanceFunc func(userId uint, amount int64) (int64, error)
	executeTransactionFunc func(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error)
}

func (m *mockWalletRepo) Get(userId uint) (*domain.Wallet, error) {
	return m.getFunc(userId)
}

func (m *mockWalletRepo) Create(wallet domain.Wallet) error {
	return m.createFunc(wallet)
}

func (m *mockWalletRepo) IncrementBalance(userId uint, amount int64) (int64, error){
	return m.incrementBalanceFunc(userId, amount)
}

func (m *mockWalletRepo) DecrementBalance(userId uint, amount int64) (int64, error){
	return m.decrementBalanceFunc(userId, amount)
}

func (m *mockWalletRepo) ExecuteTransaction(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
	return m.executeTransactionFunc(fn)
}


func TestBalance(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	t.Run("success", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return &domain.Wallet{Balance: 100000}, nil
			},
		}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(repo, txRepo, testLog)

		balance, err := service.Balance(1)

		assert.NoError(t, err)
		assert.Equal(t, "1000.00 THB", balance)
	})
	t.Run("wallet record not found", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return nil, domain.ErrNotFoundWallet
			},
		}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(repo, txRepo, testLog)
		_, err := service.Balance(999)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
	})
	t.Run("internal server error", func(t *testing.T) {
		repo := &mockWalletRepo{
			getFunc: func(id uint) (*domain.Wallet, error) {
				return nil, domain.ErrInternalServerError
			},
		}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(repo, txRepo, testLog)
		_, err := service.Balance(1)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
	})
}
func TestTopUp(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	t.Run("success topup with transaction", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		called := false
		//1. Get wallet id
		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}

		//2. Create tx 'PENDING'
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			return nil
		}

		//3. Execute db transaction
		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			return fn(walletRepo, txRepo)
		}

		//4. Increment balance
		walletRepo.incrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			return 150000, nil
		}

		//5. update status to "SUCCESS"
		txRepo.updateTxFunc = func(id uint, status string) (*domain.Transaction, error) {
			called = true

			assert.Equal(t, "SUCCESS", status)
			return &domain.Transaction{}, nil
		}

		transaction, balance, err := service.TopUp(1, 50000)

		assert.True(t, called, "should call update transaction status")
		assert.NoError(t, err)
		assert.NotNil(t, transaction)

		assert.Equal(t, float64(150000), balance)
	})
	t.Run("get wallet id failure", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		//For checking func call
		called := false

		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return nil, domain.ErrNotFoundWallet
		}

		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			called = true
			return fn(walletRepo, txRepo)
		}
		
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			called = true
			return nil
		}

		walletRepo.incrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			called = true
			return 150000, nil
		}

		transaction, balance, err := service.TopUp(1, 50000)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
		assert.Nil(t, transaction)
		assert.Equal(t, float64(0), balance)
		assert.False(t, called, "should not call db transaction")
	})
	t.Run("create transaction failure", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		//For checking func call
		called := false

		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}
		
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			return domain.ErrInternalServerError
		}

		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			called = true
			return fn(walletRepo, txRepo)
		}

		walletRepo.incrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			called = true
			return 150000, nil
		}

		transaction, balance, err := service.TopUp(1, 50000)

		assert.Error(t, err)
		assert.False(t, called, "should not call db transaction")
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.Nil(t, transaction)
		assert.Equal(t, float64(0), balance)
	})
	t.Run("update transaction status failure", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)
		
		
		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			tx.ID = 99
			return nil
		}
		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			return fn(walletRepo, txRepo)
		}
		walletRepo.incrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			return 1, nil
		}
		txRepo.updateTxFunc = func(id uint, status string) (*domain.Transaction, error) {
			return nil, domain.ErrInternalServerError
		}
		transaction, balance, err := service.TopUp(1, 50000)

		assert.NotNil(t, transaction)
		assert.Equal(t, float64(0), balance)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
	})
	t.Run("increment balance failed: transaction record not found", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)
		
		called := false

		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			tx.ID = 99
			return nil
		}
		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			return fn(walletRepo, txRepo)
		}
		walletRepo.incrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			return 0, domain.ErrNotFoundWallet
		}
		txRepo.updateTxFunc = func(id uint, status string) (*domain.Transaction, error) {
			called = true
			assert.Equal(t, "FAILED", status)
			return &domain.Transaction{Status: "FAILED"}, nil
		}
		transaction, balance, err := service.TopUp(1, 50000)


		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
		assert.Equal(t, "FAILED", transaction.Status)
		assert.Equal(t, float64(0), balance)
		assert.True(t, called, "Should call update status with 'FAILED'")
	})
}
func TestWithdraw(t *testing.T){
	testLog := logger.NewTestLogger(t)
	t.Run("success topup with transaction", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		called := false
		//1. Get wallet id
		walletRepo.getFunc = func(userID uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}

		//2. Create tx 'PENDING'
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			return nil
		}

		//3. Execute db transaction
		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			return fn(walletRepo, txRepo)
		}

		//4. Decrement balance
		walletRepo.decrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			return 150000, nil
		}

		//5. update status to "SUCCESS"
		txRepo.updateTxFunc = func(id uint, status string) (*domain.Transaction, error) {
			called = true

			assert.Equal(t, "SUCCESS", status)
			return &domain.Transaction{}, nil
		}

		transaction, balance, err := service.Withdraw(1, 50000)

		assert.True(t, called, "should call update transaction status")
		assert.NoError(t, err)
		assert.NotNil(t, transaction)

		assert.Equal(t, float64(150000), balance)
	})
	t.Run("get wallet id failure", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		//For checking func call
		called := false

		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return nil, domain.ErrNotFoundWallet
		}

		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			called = true
			return fn(walletRepo, txRepo)
		}
		
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			called = true
			return nil
		}

		walletRepo.decrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			called = true
			return 150000, nil
		}

		transaction, balance, err := service.Withdraw(1, 50000)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
		assert.Nil(t, transaction)
		assert.Equal(t, float64(0), balance)
		assert.False(t, called, "should not call db transaction")
	})
	t.Run("create transaction failure", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)

		//For checking func call
		called := false

		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}
		
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			return domain.ErrInternalServerError
		}

		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			called = true
			return fn(walletRepo, txRepo)
		}

		walletRepo.decrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			called = true
			return 150000, nil
		}

		transaction, balance, err := service.Withdraw(1, 50000)

		assert.Error(t, err)
		assert.False(t, called, "should not call db transaction")
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.Nil(t, transaction)
		assert.Equal(t, float64(0), balance)
	})
	t.Run("withdraw failed: insufficient balance for this transaction", func(t *testing.T) {
		walletRepo := &mockWalletRepo{}
		txRepo := &mockTransactionRepo{}
		service := NewWalletService(walletRepo, txRepo, testLog)
		
		
		walletRepo.getFunc = func(userId uint) (*domain.Wallet, error) {
			return &domain.Wallet{}, nil
		}
		txRepo.createTxFunc = func(tx domain.Transaction) error {
			tx.ID = 99
			return nil
		}
		walletRepo.executeTransactionFunc = func(fn func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error)) (*domain.Transaction, float64, error) {
			return fn(walletRepo, txRepo)
		}
		walletRepo.decrementBalanceFunc = func(walletId uint, amount int64) (int64, error) {
			return 0, domain.ErrInsufficientBalance
		}
		txRepo.updateTxFunc = func(id uint, status string) (*domain.Transaction, error) {
			assert.Equal(t, "FAILED", status)
			return &domain.Transaction{Status: "FAILED"}, nil
		}
		transaction, balance, err := service.Withdraw(1, 50000)

		assert.Equal(t, float64(0), balance)
		assert.NotNil(t, transaction)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInsufficientBalance, err)
	})
}
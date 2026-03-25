package repository

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormWalletRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
	
	repo := NewGormWalletRepository(gormDB)

	t.Run("success", func(t *testing.T) {
		userId := 1
		balance := 0
		currency := "THB"
		
		//Setup expectation
		mock.ExpectBegin()
	
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "wallets"`)). 
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), balance, currency, userId).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(domain.Wallet{UserID: uint(userId), Balance: int64(balance), Currency: currency})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("user already has wallet", func(t *testing.T) {
		userId := 1
		
		//Setup expectation
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "wallets"`)). 
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(0), "THB", userId).
			WillReturnError(gorm.ErrDuplicatedKey)
		mock.ExpectRollback()

		err := repo.Create(domain.Wallet{UserID: uint(userId), Balance: int64(0), Currency: "THB"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrConflictUserWallet)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("user record not found", func(t *testing.T) {
		userId := 999
		
		//Setup expectation
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "wallets"`)). 
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(0), "THB", userId).
			WillReturnError(gorm.ErrForeignKeyViolated)
		mock.ExpectRollback()

		err := repo.Create(domain.Wallet{UserID: uint(userId), Balance: 0, Currency: "THB"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFoundUser)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormWalletRepository_Get(t *testing.T){
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
	
	repo := NewGormWalletRepository(gormDB)
	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id","balance"}).AddRow(1, 100000)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id","balance" FROM "wallets" WHERE user_id = $1 AND "wallets"."deleted_at" IS NULL`)).
			WithArgs(1).
			WillReturnRows(rows)

		wallet, err := repo.Get(1)

		assert.NoError(t, err)
		assert.Equal(t, int64(100000), wallet.Balance)
		assert.Equal(t, uint(1), wallet.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("wallet record not found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id","balance"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id","balance" FROM "wallets" WHERE user_id = $1 AND "wallets"."deleted_at" IS NULL`)).
		WithArgs(999).
		WillReturnRows(rows)
		
		_, err := repo.Get(999)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Internal DB error", func(t *testing.T) {

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "id","balance" FROM "wallets" WHERE user_id = $1 AND "wallets"."deleted_at" IS NULL`)).
		WithArgs(999).
		WillReturnError(domain.ErrInternalServerError)

		_, err := repo.Get(999)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormWalletRepository_IncrementBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
	
	repo := NewGormWalletRepository(gormDB)
	t.Run("update successful", func(t *testing.T) {
		userId := uint(1)
		amount := int64(250000)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance + $1,"updated_at"=$2 WHERE user_id = $3 AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(250000))
		mock.ExpectCommit()

		balance, err := repo.IncrementBalance(userId, amount)
		
		assert.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, amount, balance)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("wallet record not found", func(t *testing.T) {
		userId := uint(999)
		amount := int64(99999)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance + $1,"updated_at"=$2 WHERE user_id = $3 AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}))
		mock.ExpectCommit()

		balance, err := repo.IncrementBalance(userId, amount)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.Equal(t, domain.ErrNotFoundWallet, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("internal server error", func(t *testing.T) {
		userId := uint(1)
		amount := int64(100000)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance + $1,"updated_at"=$2 WHERE user_id = $3 AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId).
				WillReturnError(domain.ErrInternalServerError)
		mock.ExpectRollback()

		balance, err := repo.IncrementBalance(userId, amount)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormWalletRepository_DecrementBalance(t *testing.T){
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
	
	repo := NewGormWalletRepository(gormDB)

	t.Run("update successful", func(t *testing.T) {
		userId := uint(1)
		amount := int64(250000)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance - $1,"updated_at"=$2 WHERE (user_id = $3 AND balance >= $4) AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId, amount).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(0))
		mock.ExpectCommit()

		balance, err := repo.DecrementBalance(userId, amount)
		
		assert.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, int64(0), balance)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("wallet record not found", func(t *testing.T) {
		userId := uint(999)
		amount := int64(99999)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance - $1,"updated_at"=$2 WHERE (user_id = $3 AND balance >= $4) AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId, amount).
				WillReturnRows(sqlmock.NewRows([]string{"balance"}))
		mock.ExpectCommit()

		balance, err := repo.DecrementBalance(userId, amount)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.Equal(t, domain.ErrInsufficientBalance, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("internal server error", func(t *testing.T) {
		userId := uint(1)
		amount := int64(100000)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`UPDATE "wallets" SET "balance"=balance - $1,"updated_at"=$2 WHERE (user_id = $3 AND balance >= $4) AND "wallets"."deleted_at" IS NULL RETURNING "balance"`)).
				WithArgs(amount, sqlmock.AnyArg(), userId, amount).
				WillReturnError(domain.ErrInternalServerError)
		mock.ExpectRollback()

		balance, err := repo.DecrementBalance(userId, amount)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormWalletRepository_ExecuteTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
    repo := NewGormWalletRepository(gormDB)

	t.Run("should commit when function returns value", func(t *testing.T) {
 		mock.ExpectBegin()
		mock.ExpectCommit()
		transaction, balance, err := repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {
			return &domain.Transaction{ReferenceID: "refId"}, 1000, nil
		})

		assert.NoError(t, err)
		assert.NotNil(t, balance)
		assert.NotNil(t, transaction)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("should rollback when function returns error", func(t *testing.T) {
        mock.ExpectBegin()
        mock.ExpectRollback()

        _, _, err := repo.ExecuteTransaction(func(txWallet domain.WalletRepository, txTrans domain.TransactionRepository) (*domain.Transaction, float64, error) {
			return nil, 0, errors.New("something went wrong") 
		})

        assert.Error(t, err)
        assert.Equal(t, "something went wrong", err.Error())
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}
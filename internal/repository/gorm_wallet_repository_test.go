package repository

import (
	"piano/e-wallet/internal/domain"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGormWalletRepository_CreateWallet(t *testing.T) {
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

		err := repo.CreateWallet(domain.Wallet{UserID: uint(userId), Balance: int64(balance), Currency: currency})
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

		err := repo.CreateWallet(domain.Wallet{UserID: uint(userId), Balance: int64(0), Currency: "THB"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrDuplicatedKey)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("user not found for own wallet", func(t *testing.T) {
		userId := 999
		
		//Setup expectation
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "wallets"`)). 
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), int64(0), "THB", userId).
			WillReturnError(gorm.ErrForeignKeyViolated)
		mock.ExpectRollback()

		err := repo.CreateWallet(domain.Wallet{UserID: uint(userId), Balance: 0, Currency: "THB"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrForeignKeyViolated)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormWalletRepository_GetBalance(t *testing.T){
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
		rows := sqlmock.NewRows([]string{"balance"}).AddRow(100000)
		expectedSQL := `^SELECT "balance" FROM "wallets" WHERE user_id = \$1.*`
		mock.ExpectQuery(expectedSQL).
			WithArgs(1).
			WillReturnRows(rows)

		balance, err := repo.GetBalance(1)

		assert.NoError(t, err)
		assert.Equal(t, int64(100000), balance)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid user id", func(t *testing.T) {
		expectedSQL := `^SELECT "balance" FROM "wallets" WHERE user_id = \$1.*`
		rows := sqlmock.NewRows([]string{"balance"})
		mock.ExpectQuery(expectedSQL).WillReturnRows(rows)
		
		balance, err := repo.GetBalance(999)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.EqualError(t, err, "Invalid user id")
	})
	t.Run("Internal DB error", func(t *testing.T) {
		expectedSQL := `^SELECT "balance" FROM "wallets" WHERE user_id = \$1.*`
		mock.ExpectQuery(expectedSQL).WillReturnError(domain.ErrInternalServerError)
		
		balance, err := repo.GetBalance(999)

		assert.Error(t, err)
		assert.Equal(t, int64(0), balance)
		assert.EqualError(t, err, "Internal server error")
	})
}
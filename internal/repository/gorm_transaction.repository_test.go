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

func TestGormTransactionRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
    repo := NewGormTransactionRepository(gormDB)

	t.Run("success", func(t *testing.T) {
		walletId := uint(1)
		amount := 100000
		refId := "refId"


		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				nil,
				nil,
				walletId,
				amount,
				"TOPUP",
				sqlmock.AnyArg(),
				refId,
			).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(&domain.Transaction{
			DestinationID:   &walletId, 
			Amount:          amount, 
			TransactionType: "TOPUP", 
			ReferenceID:     refId,
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	
	t.Run("Internal db error", func(t *testing.T) {
		walletId := uint(1)
		amount := 100000
		refId := "refId"


		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions"`)).
			WithArgs(
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				nil,
				nil,
				walletId,
				amount,
				"TOPUP",
				sqlmock.AnyArg(),
				refId,
			).
			WillReturnError(domain.ErrInternalServerError)
		mock.ExpectRollback()

		err := repo.Create(&domain.Transaction{
			DestinationID:   &walletId, 
			Amount:          amount, 
			TransactionType: "TOPUP", 
			Status:          "SUCCESS", 
			ReferenceID:     refId,
		})

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormTransactionRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
    repo := NewGormTransactionRepository(gormDB)

	t.Run("success", func(t *testing.T) {
		txId := uint(1)
		status := "SUCCESS"

		rows := sqlmock.NewRows([]string{"id", "status"}).
			AddRow(txId, "PENDING") // สมมติสถานะเดิมคือ PENDING
		
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE id = $1`)).
			WithArgs(txId, sqlmock.AnyArg()). // $1 คือ id, $2 คือ limit (sqlmock.AnyArg())
			WillReturnRows(rows)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "transactions" SET`)).
			WithArgs(
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
				sqlmock.AnyArg(), // deleted_at
				sqlmock.AnyArg(), // source_id
				sqlmock.AnyArg(), // destination_id
				sqlmock.AnyArg(), // amount
				sqlmock.AnyArg(), // transaction_type
				status,           // status (ตัวที่เราสนใจ)
				sqlmock.AnyArg(), // reference_id
				txId,             // id (WHERE clause)
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		result, err := repo.Update(txId, status)

		// 3. Assertions
		assert.NoError(t, err)
		assert.Equal(t, status, result.Status)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("trancsaction record not found", func(t *testing.T) {
		txId := uint(1)
		status := "SUCCESS"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE id = $1`)).
				WithArgs(txId, sqlmock.AnyArg()).
				WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.Update(txId, status)

		assert.Error(t, err)
		// assert.Equal(t, domain.ErrNotFoundTransaction, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("Internal db error", func(t *testing.T) {
		txId := uint(1)
		status := "SUCCESS"

		rows := sqlmock.NewRows([]string{"id", "status"}).
					AddRow(txId, "PENDING") // สมมติสถานะเดิมคือ PENDING
				
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE id = $1`)).
			WithArgs(txId, sqlmock.AnyArg()). // $1 คือ id, $2 คือ limit (sqlmock.AnyArg())
			WillReturnRows(rows)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "transactions" SET`)).
				WithArgs(
					sqlmock.AnyArg(), // created_at
					sqlmock.AnyArg(), // updated_at
					sqlmock.AnyArg(), // deleted_at
					sqlmock.AnyArg(), // source_id
					sqlmock.AnyArg(), // destination_id
					sqlmock.AnyArg(), // amount
					sqlmock.AnyArg(), // transaction_type
					status,           // status (ตัวที่เราสนใจ)
					sqlmock.AnyArg(), // reference_id
					txId,             // id (WHERE clause)
				).
				WillReturnError(domain.ErrInternalServerError)
		mock.ExpectRollback()

		_, err := repo.Update(txId, status)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrInternalServerError, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	
}
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

func TestGormUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
	
	repo := NewGormUserRepository(gormDB)

	t.Run("success", func(t *testing.T) {
		email := "piano@example.com"
		password := "hashed_password"
		
		//Setup expectation
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO "users"`).
			WithArgs(email, password).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		_, err := repo.Create(domain.User{Email: email, Password: password})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("email is already exists", func(t *testing.T) {
		email := "piano@example.com"
		password := "hashed_password"

		//Setup expectation
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
        	WithArgs(email, password).
        	WillReturnError(gorm.ErrDuplicatedKey)
		mock.ExpectRollback()

		_, err := repo.Create(domain.User{Email: email, Password: password})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrConflictEmail)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGormUserRepository_ExecuteTransaction(t *testing.T) {
    db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
    repo := NewGormUserRepository(gormDB)

    t.Run("should commit when function returns nil", func(t *testing.T) {
        
        mock.ExpectBegin()
        mock.ExpectCommit()
        err := repo.ExecuteTransaction(func(u domain.UserRepository, w domain.WalletRepository) error {
            return nil
        })
        assert.NoError(t, err)
        
		assert.NoError(t, mock.ExpectationsWereMet())
    })

    t.Run("should rollback when function returns error", func(t *testing.T) {
        mock.ExpectBegin()
        mock.ExpectRollback()

        err := repo.ExecuteTransaction(func(u domain.UserRepository, w domain.WalletRepository) error {
            return errors.New("something went wrong") 
        })

        assert.Error(t, err)
        assert.Equal(t, "something went wrong", err.Error())
        assert.NoError(t, mock.ExpectationsWereMet())
    })
}

func TestGormUserRepository_Find(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
	t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil{
		t.Fatalf("an error '%s' was not expected when opening a gorm database connection", err)
	}
    repo := NewGormUserRepository(gormDB)
	t.Run("find user success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "piano@example.com", "hashed_password")

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`) + `.*`).
			WithArgs("piano@example.com", 1).
			WillReturnRows(rows)
		user, err := repo.Find("piano@example.com")
		

		assert.NoError(t, err)
        assert.NotNil(t, user)
        assert.Equal(t, uint(1), user.ID)
        assert.Equal(t, "piano@example.com", user.Email)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("faliure", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`) + `.*`).
			WithArgs("piano@example.com", 1).
			WillReturnError(gorm.ErrRecordNotFound)
		user, err := repo.Find("piano@example.com")
		
		assert.Error(t, err)
        assert.Nil(t, user)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
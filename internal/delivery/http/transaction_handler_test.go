package http

import (
	"net/http/httptest"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactionService struct{
	mock.Mock
}

func (m * MockTransactionService) GetTransaction(refId string) (*domain.Transaction, error) {
	args := m.Called(refId)

	var tx *domain.Transaction
    if args.Get(0) != nil {
        tx = args.Get(0).(*domain.Transaction)
    }

	return tx, args.Error(1)
}
func (m * MockTransactionService) GetAllTransaction(userId uint) ([]domain.Transaction, error) {
	args := m.Called(userId)

	var transactions []domain.Transaction
    if args.Get(0) != nil {
        transactions = args.Get(0).([]domain.Transaction)
    }

	return transactions, args.Error(1)
}

func TestGetTransactionHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService, testLog)

	app := fiber.New()
	app.Get("/transaction/:refId", handler.GetTransaction)

	t.Run("success", func(t *testing.T) {
		refId := "REF_ID"
		userId := uint(1)
		app.Get("/transaction/:refId", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetTransaction)
		
		mockService.On("GetTransaction", refId).
			Return(&domain.Transaction{ReferenceID: refId}, nil).
			Once()

		req := httptest.NewRequest("GET", "/transaction/"+refId, nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: refId param is missing", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		app.Get("/transaction/:refId?", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetTransaction)
		
		req := httptest.NewRequest("GET", "/transaction/", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.ErrBadRequest.Code, resp.StatusCode)
		mockService.AssertNotCalled(t, "GetTransaction", "")
		mockService.AssertExpectations(t)
	})
	t.Run("failure: transaction not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		refId := "REF_ID"
		userId := uint(1)
		app.Get("/transaction/:refId", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetTransaction)

		mockService.On("GetTransaction", refId).
			Return(nil, domain.ErrNotFoundTransaction).
			Once()

		req := httptest.NewRequest("GET", "/transaction/"+refId, nil)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.ErrNotFound.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		refId := "REF_ID"
		userId := uint(1)
		app.Get("/transaction/:refId", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetTransaction)

		mockService.On("GetTransaction", refId).
			Return(nil, domain.ErrInternalServerError).
			Once()

		req := httptest.NewRequest("GET", "/transaction/"+refId, nil)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.ErrInternalServerError.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestGetAllTransactionHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService, testLog)

	t.Run("success", func(t *testing.T) {
		userId := uint(1)
		expectedTransactions := []domain.Transaction{
			{TransactionType: "TOPUP"},
			{TransactionType: "WITHDRAW"},
		}
		app := fiber.New()
		app.Get("/transaction", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetAllTransaction)

		mockService.On("GetAllTransaction", userId).Return(expectedTransactions, nil)

		req := httptest.NewRequest("GET", "/transaction", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: user context is missing", func(t *testing.T) {
		app := fiber.New()
		app.Get("/transaction", func(c fiber.Ctx) error {
					c.Locals("userId")
					return c.Next()
				}, handler.GetAllTransaction)

		req := httptest.NewRequest("GET", "/transaction", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode)
		mockService.AssertNotCalled(t, "GetAllTransaction", "")
		mockService.AssertExpectations(t)
	})
	t.Run("failure: wallet record not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		app := fiber.New()
		app.Get("/transaction", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetAllTransaction)

		mockService.On("GetAllTransaction", userId).Return(nil, domain.ErrNotFoundWallet)

		req := httptest.NewRequest("GET", "/transaction", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		app := fiber.New()
		app.Get("/transaction", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.GetAllTransaction)

		mockService.On("GetAllTransaction", userId).Return(nil, domain.ErrInternalServerError)

		req := httptest.NewRequest("GET", "/transaction", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
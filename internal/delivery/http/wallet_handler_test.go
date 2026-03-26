package http

import (
	"bytes"
	"io"
	"net/http/httptest"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


type MockWalletService struct{
	mock.Mock
}

func (m *MockWalletService) Balance(userId uint) (string, error) {
	args := m.Called(userId)
	return args.String(0), args.Error(1)
}

func (m *MockWalletService) TopUp(userId uint, amount float64) (*domain.Transaction, float64, error){
	args := m.Called(userId, amount)

    var tx *domain.Transaction
    if args.Get(0) != nil {
        tx = args.Get(0).(*domain.Transaction)
    }

    balance := args.Get(1).(float64)

    return tx, balance, args.Error(2)
}

func (m *MockWalletService) Withdraw(userId uint, amount float64) (*domain.Transaction, float64, error){
	args := m.Called(userId, amount)

    var tx *domain.Transaction
    if args.Get(0) != nil {
        tx = args.Get(0).(*domain.Transaction)
    }

    balance := args.Get(1).(float64)

    return tx, balance, args.Error(2)
}

func (m *MockWalletService) Transfer(userId uint, desId uint, amount float64) (*domain.Transaction, float64, error){
	args := m.Called(userId, desId, amount)

    var tx *domain.Transaction
    if args.Get(0) != nil {
        tx = args.Get(0).(*domain.Transaction)
    }

    balance := args.Get(1).(float64)

    return tx, balance, args.Error(2)
}

func (m *MockWalletService) Info(userId uint) (uint, error) {
	args := m.Called(userId)
	id := args.Get(0).(uint)
	return id, args.Error(1)
}

func TestBalanceHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService, testLog)

	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userId := uint(1)

		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)

		mockService.On("Balance", userId).Return("1000.00 THB", nil)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), "1000.00 THB")
		mockService.AssertExpectations(t)
	})
	t.Run("wallet record not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := uint(999)
		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)
		mockService.On("Balance", userId).Return("", domain.ErrNotFoundWallet)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("Internal DB error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := uint(1)
		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)
		mockService.On("Balance", userId).Return("", domain.ErrInternalServerError)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestTopUpHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService, testLog)

	t.Run("topup success", func(t *testing.T) {
		userId := uint(1)
		amount := float64(1000)
		app := fiber.New()
		app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.TopUp)
		mockService.On("TopUp", userId, amount).Return(&domain.Transaction{}, amount, nil)

		req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("missing user context", func(t *testing.T) {

	app := fiber.New()
	app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId")
					return c.Next()
				}, handler.TopUp)
	req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")
    resp, _ := app.Test(req)

    assert.Equal(t, 500, resp.StatusCode)
    mockService.AssertNotCalled(t, "TopUp", "", mock.Anything)
	})
	t.Run("wallet not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(999)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.TopUp)
		mockService.On("TopUp", userId, amount).Return(nil, float64(0), domain.ErrNotFoundWallet)

		req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("Invalid req format", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(999)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.TopUp)
		mockService.On("TopUp", userId, amount).Return(nil, float64(0), fiber.StatusBadRequest)

		req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 1000`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("amount required", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.TopUp)
		mockService.On("TopUp", userId, amount).Return(nil, float64(0), fiber.StatusBadRequest)

		req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 0}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, string(body), "Amount is required")
		mockService.AssertExpectations(t)
	})
	t.Run("internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/topup", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.TopUp)
		mockService.On("TopUp", userId, amount).Return(nil, float64(0), domain.ErrInternalServerError)

		req := httptest.NewRequest("POST", "/topup", bytes.NewBufferString(`{"amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestWithdraw(t *testing.T){
	testLog := logger.NewTestLogger(t)
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService, testLog)

	t.Run("withdraw success", func(t *testing.T) {
		userId := uint(1)
		amount := float64(1000)
		app := fiber.New()
		app.Post("/withdraw", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Withdraw)
		mockService.On("Withdraw", userId, amount).Return(&domain.Transaction{}, float64(0), nil)

		req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("missing user context", func(t *testing.T) {

	app := fiber.New()
	app.Post("/withdraw", func(c fiber.Ctx) error {
				c.Locals("userId")
				return c.Next()
			}, handler.Withdraw)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)

	assert.NoError(t, err)
    assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
    mockService.AssertNotCalled(t, "Withdraw", "", mock.Anything)
	})
	t.Run("Invalid req format", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)

		app := fiber.New()
		app.Post("/withdraw", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Withdraw)
		
		req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 0`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("Validation error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)

		app := fiber.New()
		app.Post("/withdraw", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Withdraw)
		
		req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 0`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("insufficient balance", func(t *testing.T) {
	mockService.ExpectedCalls = nil
	userId := uint(1)
	amount := float64(1000)
	app := fiber.New()
	app.Post("/withdraw", func(c fiber.Ctx) error {
				c.Locals("userId", userId)
				return c.Next()
			}, handler.Withdraw)
	mockService.On("Withdraw", userId, amount).Return(&domain.Transaction{Status: "FAILED"}, float64(0), domain.ErrInsufficientBalance)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	mockService.AssertExpectations(t)
	})
	t.Run("transaction is already created", func(t *testing.T) {
	mockService.ExpectedCalls = nil
	userId := uint(1)
	amount := float64(1000)
	app := fiber.New()
	app.Post("/withdraw", func(c fiber.Ctx) error {
				c.Locals("userId", userId)
				return c.Next()
			}, handler.Withdraw)
	mockService.On("Withdraw", userId, amount).Return(&domain.Transaction{Status: "FAILED"}, float64(0), domain.ErrConflictTransactionRefId)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
	mockService.AssertExpectations(t)
	})
	t.Run("update transaction: transaction record not found", func(t *testing.T) {
	mockService.ExpectedCalls = nil
	userId := uint(1)
	amount := float64(1000)
	app := fiber.New()
	app.Post("/withdraw", func(c fiber.Ctx) error {
				c.Locals("userId", userId)
				return c.Next()
			}, handler.Withdraw)
	mockService.On("Withdraw", userId, amount).Return(&domain.Transaction{Status: "FAILED"}, float64(0), domain.ErrNotFoundTransaction)

	req := httptest.NewRequest("POST", "/withdraw", bytes.NewBufferString(`{"amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	mockService.AssertExpectations(t)
	})
}

func TestTransfer(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService, testLog)

	t.Run("success", func(t *testing.T) {
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(&domain.Transaction{TransactionType: "TRANSFER"}, float64(0), nil)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("missing user context", func(t *testing.T) {

	app := fiber.New()
	app.Post("/transfer", func(c fiber.Ctx) error {
				c.Locals("userId")
				return c.Next()
			}, handler.Transfer)

	req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
	req.Header.Set("Content-Type", "application/json")
    resp, err := app.Test(req)

	assert.NoError(t, err)
    assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
    mockService.AssertNotCalled(t, "Transfer", mock.AnythingOfType("string"), mock.AnythingOfType("uint"), mock.AnythingOfType("float64"))
	mockService.AssertExpectations(t)
	})
	t.Run("invalid req format", func(t *testing.T) {
		userId := uint(1)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertNotCalled(t, "Transfer", mock.AnythingOfType("string"), mock.AnythingOfType("uint"), mock.AnythingOfType("float64"))
		mockService.AssertExpectations(t)
	})
	t.Run("validation error: required", func(t *testing.T) {
		userId := uint(1)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertNotCalled(t, "Transfer", mock.AnythingOfType("string"), mock.AnythingOfType("uint"), mock.AnythingOfType("float64"))
		mockService.AssertExpectations(t)
	})
	t.Run("can't transfer to own wallet", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrConflictSourceDesId)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("wallet not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrNotFoundWallet)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.ErrNotFound.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("insufficient balance for this transaction", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrInsufficientBalance)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("this transaction is already created", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrConflictTransactionRefId)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("transaction record not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrNotFoundTransaction)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		userId := uint(1)
		desId := uint(2)
		amount := float64(1000)

		app := fiber.New()
		app.Post("/transfer", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Transfer)
		mockService.On("Transfer", userId, desId, amount).Return(nil, float64(0), domain.ErrInternalServerError)

		req := httptest.NewRequest("POST", "/transfer", bytes.NewBufferString(`{"destination_id": 2, "amount": 1000}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}

func TestInfo(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService, testLog)
	
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userId := uint(1)
		walletId := uint(2)

		app.Get("/info", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Info)

		mockService.On("Info", userId).Return(walletId, nil)

		req := httptest.NewRequest("GET", "/info", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: missing user context", func(t *testing.T) {
		app := fiber.New()
		app.Get("/info", func(c fiber.Ctx) error {
					c.Locals("userId")
					return c.Next()
				}, handler.Info)

		req := httptest.NewRequest("GET", "/info", nil)
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertNotCalled(t, "Info", mock.AnythingOfType("sting"))
		mockService.AssertExpectations(t)
	})
	t.Run("failure: wallet not found", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := uint(1)

		app.Get("/info", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Info)

		mockService.On("Info", userId).Return(uint(0), domain.ErrNotFoundWallet)

		req := httptest.NewRequest("GET", "/info", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("failure: internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := uint(1)

		app.Get("/info", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Info)

		mockService.On("Info", userId).Return(uint(0), domain.ErrInternalServerError)

		req := httptest.NewRequest("GET", "/info", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
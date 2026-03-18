package http

import (
	"io"
	"net/http/httptest"
	"piano/e-wallet/internal/domain"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


type MockWalletService struct{
	mock.Mock
}

func (m * MockWalletService) Balance(userId uint) (string, error) {
	args := m.Called(userId)
	return args.String(0), args.Error(1)
}

func TestBalanceHandler(t *testing.T) {
	mockService := new(MockWalletService)
	handler := NewWalletHandler(mockService)

	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userId := 1
		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)
		mockService.On("Balance", uint(userId)).Return("1000.00 THB", nil)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Contains(t, string(body), "1000.00 THB")
		mockService.AssertExpectations(t)
	})

	t.Run("Invalid user id", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := 999
		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)
		mockService.On("Balance", uint(userId)).Return("", domain.ErrUserRecordNotFound)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
		assert.Contains(t, string(body), "Invalid user id")
		mockService.AssertExpectations(t)
	})
	t.Run("Internal DB error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		app := fiber.New()
		userId := 1
		app.Get("/balance", func(c fiber.Ctx) error {
					c.Locals("userId", userId)
					return c.Next()
				}, handler.Balance)
		mockService.On("Balance", uint(userId)).Return("", domain.ErrInternalServerError)

		req := httptest.NewRequest("GET", "/balance", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
		assert.Contains(t, string(body), "Internal server error")
		mockService.AssertExpectations(t)
	})
}
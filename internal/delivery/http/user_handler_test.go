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

type MockUserService struct{
	mock.Mock
}

func (m * MockUserService) Register(user domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestRegisterHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService, testLog)

	app := fiber.New()
	app.Post("/register", handler.Register)

	t.Run("success user creation", func(t *testing.T) {
		mockService.On("Register", mock.AnythingOfType("domain.User")).Return(nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(`{"email": "piano@example.com", "password": "password"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register",  bytes.NewBufferString(`{"email": "piano@example.com", "password": "password"`))

		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("validation error cases", func(t *testing.T) {
		tests := []struct{
			name string
			requestBody string
			expectedStatus int
			expectedError string
		}{
			{
				name: "missing email",
				requestBody: `{"password": "password"}`,
				expectedStatus: fiber.StatusBadRequest,
				expectedError: "Email is required",
			},
			{
				name: "invalid email format",
				requestBody: `{"email": "piano@example", "password": "password"}`,
				expectedStatus: fiber.StatusBadRequest,
				expectedError: "invalid email format",
			},
			{
				name: "password too short",
				requestBody: `{"password": "pass"}`,
				expectedStatus: fiber.StatusBadRequest,
				expectedError: "Password is must be 8 character",
			},
		}

		for _, tt := range tests{
			req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(tt.requestBody))

			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
            
            body, _ := io.ReadAll(resp.Body)
            assert.Contains(t, string(body), tt.expectedError)
		}
	})

	t.Run("email is already exist", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		mockService.On("Register", mock.Anything).Return(domain.ErrConflictEmail)

		req := httptest.NewRequest("POST", "/register",  bytes.NewBufferString(`{"email": "piano@example.com", "password": "password"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusConflict, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
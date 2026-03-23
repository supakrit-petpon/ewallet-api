package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/pkg/logger"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct{
	mock.Mock
}

func (m * MockAuthService) Login(email string, password string) (string, error) {
	args := m.Called(email, password)
	return args.String(0), args.Error(1)
}

func TestLoginHandler(t *testing.T) {
	testLog := logger.NewTestLogger(t)
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService, testLog)

	app := fiber.New()
	app.Post("/login", handler.Login)
	t.Run("login success", func(t *testing.T) {
		mockService.On("Login", "piano@example.com", "valid_password").Return("token", nil)

		reqBody := `{"email":"piano@example.com", "password":"valid_password"}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		// เช็ค Cookie
		cookies := resp.Cookies()
		var jwtCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "jwt" {
				jwtCookie = cookie
				break
			}
		}

		assert.NotNil(t, jwtCookie, "Should have jwt cookie")
		assert.Equal(t, "token", jwtCookie.Value)
		assert.True(t, jwtCookie.HttpOnly)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request", func(t *testing.T) {
		reqBody := `{"email":"piano@example.com", "password":"valid_password"` //ไม่ได้ใส่ "}"
		req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		body, _ := io.ReadAll(resp.Body)
		
		assert.Contains(t, string(body), "invalid request")
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
				name: "missing password",
				requestBody: `{"email": "piano@example.com"}`,
				expectedStatus: fiber.StatusBadRequest,
				expectedError: "Password is required",
			},
			{
				name: "invalid email format",
				requestBody: `{"email": "piano@example"}`,
				expectedStatus: fiber.StatusBadRequest,
				expectedError: "invalid email format",
			},

		}

		for _, tt := range tests{
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tt.requestBody))

			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
            
            body, _ := io.ReadAll(resp.Body)
            assert.Contains(t, string(body), tt.expectedError)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		mockService.On("Login", "piano@example.com", "valid_password").Return("", domain.ErrAuthUnauthorized)

		reqBody := `{"email":"piano@example.com", "password":"valid_password"}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.ErrUnauthorized.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		mockService.On("Login", "piano@example.com", "wrong_password").Return("", domain.ErrAuthUnauthorized)

		reqBody := `{"email":"piano@example.com", "password":"wrong_password"}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.ErrUnauthorized.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
	t.Run("internal server error", func(t *testing.T) {
		mockService.ExpectedCalls = nil
		mockService.On("Login", "piano@example.com", "valid_password").Return("", domain.ErrInternalServerError)

		reqBody := `{"email":"piano@example.com", "password":"valid_password"}`
		req := httptest.NewRequest("POST", "/login", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.ErrInternalServerError.Code, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
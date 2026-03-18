package middleware

import (
	"io"
	"net/http/httptest"
	"piano/e-wallet/internal/app"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"piano/e-wallet/internal/infrastructure/jwt"

	jwtv4 "github.com/golang-jwt/jwt/v4"
)

func TestAuthRequired(t *testing.T) {
	
	secret := "test-secret-key"
	cfg := &app.Application{
		Config: &app.Config{
			SecretKey: secret,
		},
	}

	app := fiber.New()
	app.Use(AuthRequired(cfg))

	app.Get("/balance", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"balance": "100.00 THB",
		})
	})
		//Setup Mock Config
		handler := &jwt.JWTHandler{SecretKey: secret}
		userId := uint(123)

		//Mock Invalid token
		tokenNoUser := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, jwtv4.MapClaims{
        "exp": time.Now().Add(time.Hour).Unix(),
        // ไม่ใส่ user_id
    	})
    	tokenNoUserStr, _ := tokenNoUser.SignedString([]byte(secret))

		tokenWrongType := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, jwtv4.MapClaims{
        "user_id": "123", // ส่งเป็น string แทนตัวเลข
        "exp":     time.Now().Add(time.Hour).Unix(),
    	})
    	tokenWrongTypeStr, _ := tokenWrongType.SignedString([]byte(secret))
	
		//Test Cases
	tests := []struct{
		description string
		route string
		httpMethod string
		token string
		expectedCode uint
		expectedBody string
	}{
		{
			description: "Valid token, access granted",
			route: "/balance",
			httpMethod: "GET",
			token: must(handler.GenerateToken(userId)),
			expectedCode: fiber.StatusOK,
			expectedBody: `{"balance":"100.00 THB"}`,
		},
		{
			description: "Invalid token format, access deniedd",
			route: "/balance",
			httpMethod: "GET",
			token: "invalidtoken",
			expectedCode: fiber.StatusUnauthorized,
			expectedBody: "",
		},
		{
			description: "Missing token, access denied",
			route: "/balance",
			httpMethod: "GET",
			token: "",
			expectedCode: fiber.StatusUnauthorized,
			expectedBody: "",
		},
		{
			description: "Missing user_id in token acess denied",
			route: "/balance",
			httpMethod: "GET",
			token: tokenNoUserStr,
			expectedCode: fiber.StatusUnauthorized,
			expectedBody: "",
		},
		{
			description: "Invalid type of user_id in token, access denied",
			route: "/balance",
			httpMethod: "GET",
			token: tokenWrongTypeStr,
			expectedCode: fiber.StatusUnauthorized,
			expectedBody: "",
		},
	}


	for _, test := range tests{
		t.Run(test.description, func(t *testing.T) {
			req := httptest.NewRequest(test.httpMethod, test.route, nil)
			req.Header.Set("Cookie", "jwt=" + test.token)
			
			resp, err := app.Test(req, fiber.TestConfig{FailOnTimeout: false})
			assert.NoError(t, err)

			assert.Equal(t, int(test.expectedCode), resp.StatusCode, test.description)

			if test.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedBody, string(body), test.description)
			}
		})
	}
}

// Helper function to handle error from generateToken in a test environment
func must[T any](v T, err error) T {
    if err != nil {
        panic(err)
    }
    return v
}
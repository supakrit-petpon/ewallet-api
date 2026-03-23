package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v4"
)


func AuthRequired(secretKey string) fiber.Handler{
	return func(c fiber.Ctx) error{
		cookie := c.Cookies("jwt")
		if cookie == "" {
    		return c.SendStatus(fiber.StatusUnauthorized)
		}

		token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
		return c.SendStatus(fiber.StatusUnauthorized)
		}

		claims := token.Claims.(jwt.MapClaims)
		
		val, exists := claims["user_id"]
		if !exists {
			return c.Status(fiber.StatusUnauthorized).SendString("user_id is missing")
		}

		userId, ok := val.(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).SendString("invalid user_id type")
		}

		c.Locals("userId", uint(userId))
		return c.Next()
	}
}
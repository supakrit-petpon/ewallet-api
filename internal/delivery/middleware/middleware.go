package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v4"

	"piano/e-wallet/internal/app"
)


func AuthRequired(cfg *app.Application) fiber.Handler{
	return func(c fiber.Ctx) error{
		cookie := c.Cookies("jwt")
		if cookie == "" {
    		return c.SendStatus(fiber.StatusUnauthorized)
		}

		token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Config.SecretKey), nil
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

		c.Locals("userId", int(userId))
		return c.Next()
	}
}
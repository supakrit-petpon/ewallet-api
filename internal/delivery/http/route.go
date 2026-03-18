package http

import (
	"piano/e-wallet/internal/delivery/middleware"

	"github.com/gofiber/fiber/v3"
)

func MapRoutes(app *fiber.App, secretKey string, userHandler *UserHandler, authHandler *AuthHandler, walletHandler *WalletHandler) {
    // API Grouping
    v1 := app.Group("/api/v1")

    // User Routes
    users := v1.Group("/users")
    users.Post("/", userHandler.Register)

    //Auth Routes
    auth := v1.Group("/auth")
    auth.Post("/login", authHandler.Login)
    
    //Wallet Routes
    wallet := v1.Group("/wallet")
    wallet.Use(middleware.AuthRequired)
    wallet.Get("/balance", walletHandler.Balance)
}
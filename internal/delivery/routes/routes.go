package routes

import (
	"piano/e-wallet/internal/delivery/http"
	"piano/e-wallet/internal/delivery/middleware"

	"github.com/gofiber/fiber/v3"
)

func MapRoutes(app *fiber.App, secretKey string, userHandler *http.UserHandler, authHandler *http.AuthHandler, walletHandler *http.WalletHandler) {
    // API Grouping
    v1 := app.Group("/api/v1")

    // User Routes
    v1.Post("/register", userHandler.Register)

    //Auth Routes
    auth := v1.Group("/auth")
    auth.Post("/login", authHandler.Login)
    
    //Wallet Routes
    wallet := v1.Group("/wallet")
    wallet.Use(middleware.AuthRequired(secretKey))
    wallet.Get("/balance", walletHandler.Balance)
    wallet.Post("/topup", walletHandler.TopUp)
    wallet.Post("/withdraw", walletHandler.Withdraw)
}
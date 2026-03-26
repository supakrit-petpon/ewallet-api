package routes

import (
	"piano/e-wallet/internal/delivery/http"
	"piano/e-wallet/internal/delivery/middleware"

	"github.com/gofiber/fiber/v3"
)

func MapRoutes(app *fiber.App, secretKey string, userHandler *http.UserHandler,
     authHandler *http.AuthHandler, walletHandler *http.WalletHandler,
     transactionHandler *http.TransactionHandler) {
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
    wallet.Post("/transfer", walletHandler.Transfer)
    wallet.Get("/info", walletHandler.Info)

    //Transaction Routes
    transaction := v1.Group("/transaction")
    transaction.Use(middleware.AuthRequired(secretKey))
    transaction.Get("/:refId", transactionHandler.GetTransaction)
    transaction.Get("/", transactionHandler.GetAllTransaction)
}
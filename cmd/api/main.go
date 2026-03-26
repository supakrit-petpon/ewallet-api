package main

import (
	"log"
	"piano/e-wallet/internal/app"
	"piano/e-wallet/internal/delivery/http"
	"piano/e-wallet/internal/delivery/routes"
	"piano/e-wallet/internal/infrastructure/jwt"
	"piano/e-wallet/internal/repository"
	"piano/e-wallet/internal/usecases"
)

func main(){
	application := app.NewApplication()
	logger := application.Logger

	//1. Set up User Layer
	userRepo := repository.NewGormUserRepository(application.DB)
	userService := usecases.NewUserService(userRepo, logger)
	userHandler := http.NewUserHandler(userService, logger)

	//2. Auth
	tokenProvider := jwt.NewTokenProvider(application.Config.SecretKey)
	authService := usecases.NewAuthService(userRepo, tokenProvider, logger)
	authHandler := http.NewAuthHandler(authService, logger)

	//3. Transaction
	txRepo := repository.NewGormTransactionRepository(application.DB)
	txService := usecases.NewTransactionService(txRepo, logger)
	txHandler := http.NewTransactionHandler(txService, logger)

	//4. Wallet
	walletRepo := repository.NewGormWalletRepository(application.DB)
	walletService := usecases.NewWalletService(walletRepo, txRepo, logger)
	walletHandler := http.NewWalletHandler(walletService, logger)

	routes.MapRoutes(application.App, application.Config.SecretKey, userHandler, authHandler, walletHandler, txHandler)

	err := application.Start()
	if err != nil {
		log.Fatalf("Could not start application: %v", err)
	}
}
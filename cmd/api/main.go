package main

import (
	"log"
	"piano/e-wallet/internal/app"
	"piano/e-wallet/internal/delivery/http"
	"piano/e-wallet/internal/delivery/routes"
	"piano/e-wallet/internal/infrastructure/jwt"
	"piano/e-wallet/internal/repository"
	"piano/e-wallet/internal/usecases"

	_ "piano/e-wallet/docs"
)

// @title E-Wallet API
// @description This is a server for a wallet API.
// @version 1.0
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
func main(){
	application := app.NewApplication()
	logger := application.Logger

	//1. Set up Repo & Provider
	userRepo := repository.NewGormUserRepository(application.DB)
	tokenProvider := jwt.NewTokenProvider(application.Config.SecretKey)
	txRepo := repository.NewGormTransactionRepository(application.DB)
	walletRepo := repository.NewGormWalletRepository(application.DB)

	//2. Set up Service Layers
	userService := usecases.NewUserService(userRepo, logger)
	authService := usecases.NewAuthService(userRepo, tokenProvider, logger)
	txService := usecases.NewTransactionService(txRepo, walletRepo, logger)
	walletService := usecases.NewWalletService(walletRepo, txRepo, logger)
	
	//3. Set up Handler Layers
	userHandler := http.NewUserHandler(userService, logger)
	authHandler := http.NewAuthHandler(authService, logger)
	txHandler := http.NewTransactionHandler(txService, logger)
	walletHandler := http.NewWalletHandler(walletService, logger)

	routes.MapRoutes(application.App, application.Config.SecretKey, userHandler, authHandler, walletHandler, txHandler)

	err := application.Start()
	if err != nil {
		log.Fatalf("Could not start application: %v", err)
	}
}
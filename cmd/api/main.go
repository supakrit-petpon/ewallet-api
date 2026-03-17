package main

import (
	"log"
	"piano/e-wallet/internal/app"
	"piano/e-wallet/internal/delivery/http"
	"piano/e-wallet/internal/infrastructure/jwt"
	"piano/e-wallet/internal/repository"
	"piano/e-wallet/internal/usecases"
)

func main(){
	application := app.NewApplication()

	//1. Set up User Layer
	userRepo := repository.NewGormUserRepository(application.DB)
	userService := usecases.NewUserService(userRepo)
	userHandler := http.NewUserHandler(userService)

	//2. Auth
	tokenProvider := jwt.NewTokenProvider(application.Config.SecretKey)
	authService := usecases.NewAuthService(userRepo, tokenProvider)
	authHandler := http.NewAuthHandler(authService)

	http.MapRoutes(application.App, userHandler, authHandler)

	err := application.Start()
	if err != nil {
		log.Fatalf("Could not start application: %v", err)
	}
}
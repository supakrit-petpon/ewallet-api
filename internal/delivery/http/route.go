package http

import "github.com/gofiber/fiber/v3"

func MapRoutes(app *fiber.App, userHandler *UserHandler, authHandler *AuthHandler) {
    // API Grouping
    v1 := app.Group("/api/v1")

    // User Routes
    users := v1.Group("/users")
    users.Post("/", userHandler.Register)

    //Auth Routes
    auth := v1.Group("/auth")
    auth.Post("/login", authHandler.Login)
}
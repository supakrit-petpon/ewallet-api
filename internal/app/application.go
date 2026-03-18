package app

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)


type Application struct {
	Config *Config
	DB     *gorm.DB
	App *fiber.App
}

func NewApplication() *Application {
	cfg := LoadConfig()
	db := DBSetup(*cfg)
	
	router := fiber.New(fiber.Config{
		AppName: "E-wallet api",
		StrictRouting: false,
	})

	log.Println("Internal application setup complete")

	return &Application{
		Config: cfg,
		DB:     db,
		App: router,
	}
}

func (a *Application) Start() error {
	a.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})
	
	log.Printf("Application starting on port %s...", a.Config.Port)
	
	return a.App.Listen(":" + a.Config.Port)
}
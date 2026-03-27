package app

import (
	"log"
	"piano/e-wallet/pkg/logger"

	_ "piano/e-wallet/docs"

	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type Application struct {
	Config *Config
	DB     *gorm.DB
	App *fiber.App
	Logger logger.Logger
}

func NewApplication() *Application {
	cfg := LoadConfig()
	db := DBSetup(*cfg)
	logger := logger.NewZapLogger()
	
	router := fiber.New(fiber.Config{
		AppName: "E-wallet api",
		StrictRouting: false,
	})

	logger.Info("Internal application setup complete")

	return &Application{
		Config: cfg,
		DB:     db,
		App: router,
		Logger: logger,
	}
}

func (a *Application) Start() error {
	a.App.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	a.App.Get("/swagger/*", swaggo.HandlerDefault)
	
	
	log.Printf("Application starting on port %s...", a.Config.Port)
	
	return a.App.Listen(":" + a.Config.Port)
}
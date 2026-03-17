package app

import (
	"log"
	"os"
	"piano/e-wallet/internal/domain"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


type Config struct {
	Port string
	DBDSN string
	SecretKey string
}

type Application struct {
	Config Config
	DB     *gorm.DB
	App *fiber.App
}

func NewApplication() *Application {
	// 1. Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	cfg := Config{
		Port:  os.Getenv("PORT"),
		DBDSN: os.Getenv("DB_DSN"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}
	// 2. New logger for detailed SQL logging
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
		SlowThreshold: time.Second, // Slow SQL threshold
		LogLevel:      logger.Info, // Log level
		Colorful:      true,        // Enable color
		},
	)
	// 3. Connect to Database (GORM)
	db, err := gorm.Open(postgres.Open(cfg.DBDSN), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// 4. Migrate DB
	db.AutoMigrate(&domain.User{}, &domain.Wallet{})

	// 5. Init Fiber Router
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
	// คุณสามารถลงทะเบียน Middleware หรือ Routes ตรงนี้ได้

	log.Printf("Application starting on port %s...", a.Config.Port)
	
	return a.App.Listen(":" + a.Config.Port)
}
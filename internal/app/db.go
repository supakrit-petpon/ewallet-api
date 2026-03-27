package app

import (
	"fmt"
	"log"
	"os"
	"piano/e-wallet/internal/domain"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func DBSetup(cfg Config) *gorm.DB  {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
		SlowThreshold: time.Second, // Slow SQL threshold
		LogLevel:      logger.Info, // Log level
		Colorful:      true,        // Enable color
		},
	)
	dbdsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
	 cfg.DB_HOST, cfg.DB_PORT, cfg.DB_USER, cfg.DB_PASSWORD, cfg.DB_NAME)

	db, err := gorm.Open(postgres.Open(dbdsn), &gorm.Config{
		Logger: newLogger,
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	// 4. Migrate DB
	db.AutoMigrate(&domain.User{}, &domain.Wallet{}, &domain.Transaction{})

	return db
}
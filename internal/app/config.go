package app

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DBDSN string
	SecretKey string
}

func LoadConfig() *Config{
	godotenv.Load("../../.env")
	return &Config{
		Port:  os.Getenv("PORT"),	
		DBDSN: os.Getenv("DB_DSN"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}
}

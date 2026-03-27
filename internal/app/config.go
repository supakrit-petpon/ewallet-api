package app

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	DB_HOST string
	DB_PORT int
	DB_USER string
	DB_PASSWORD string
	DB_NAME string
	SecretKey string
}

func LoadConfig() *Config{
	godotenv.Load("../../.env")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil
	}
	return &Config{
		Port:  os.Getenv("PORT"),	
		DB_HOST: os.Getenv("DB_HOST"),
		DB_PORT: port,
		DB_USER: os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_NAME: os.Getenv("DB_NAME"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}
}

package config

import (
	"log"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
}

func Load() *Config {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8083"),
		DatabaseURL: dbURL,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

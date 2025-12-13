package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	AppPort    string
}

func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:     getEnv("POSTGRES_DB", "postgres"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		AppPort:    getEnv("PORT", "8000"),
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

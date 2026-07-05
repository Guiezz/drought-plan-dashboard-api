package db

import (
	"fmt"
	"log"

	"github.com/guiezz/dashboard-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func safeDSN(cfg *config.Config) string {
	return fmt.Sprintf("host=%s user=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)
}

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("[ERRO] Falha ao conectar no banco (host=%s, user=%s, dbname=%s): %v",
			cfg.DBHost, cfg.DBUser, cfg.DBName, err)
		return nil, fmt.Errorf("falha ao conectar no banco de dados")
	}

	return db, nil
}

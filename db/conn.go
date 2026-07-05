package db

import (
	"fmt"
	"log"

	"github.com/guiezz/dashboard-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

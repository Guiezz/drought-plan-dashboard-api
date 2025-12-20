package db

import (
	"fmt"

	"github.com/guiezz/dashboard-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: false,
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco de dados (DSN: %s): %w", dsn, err)
	}

	return db, nil
}

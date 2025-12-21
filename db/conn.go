package db

import (
	"fmt"

	"github.com/guiezz/dashboard-api/config" // <--- Novo Import
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mude a assinatura para receber o *config.Config
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	// Usa o método GetDSN() da configuração para obter a string de conexão
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// Mensagem de erro mais clara
		return nil, fmt.Errorf("falha ao conectar no banco de dados (DSN: %s): %w", dsn, err)
	}

	return db, nil
}

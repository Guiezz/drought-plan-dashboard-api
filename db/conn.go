package db

import (
	"fmt"
	"log"

	"github.com/guiezz/dashboard-api/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// dialector usa o protocolo simples do Postgres: sem statements preparados no
// servidor, o pooler não guarda planos que quebram quando a tabela ganha colunas
// ("cached plan must not change result type").
func dialector(dsn string) gorm.Dialector {
	return postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true})
}

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(dialector(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("[ERRO] Falha ao conectar no banco (host=%s, user=%s, dbname=%s): %v",
			cfg.DBHost, cfg.DBUser, cfg.DBName, err)
		return nil, fmt.Errorf("falha ao conectar no banco de dados")
	}

	return db, nil
}

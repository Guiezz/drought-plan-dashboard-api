package db

import (
	"context"
	"fmt"
	"net"

	"github.com/guiezz/dashboard-api/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDB conecta ao banco forçando IPv4 e desativando Prepared Statements
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// 1. Analisa a string de conexão (DSN) para criar a configuração do driver
	dbConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar configuração do banco: %w", err)
	}

	// 2. TRUQUE PARA O RENDER: Forçar conexão via IPv4 (tcp4)
	// O Render não suporta IPv6 de saída, mas o DNS do Supabase retorna IPv6.
	// Esta função obriga o Go a resolver apenas o endereço IPv4.
	dbConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return net.Dial("tcp4", addr)
	}

	// 3. Abre a conexão "low-level" usando o driver configurado
	sqlDB := stdlib.OpenDB(*dbConfig)

	// 4. Inicializa o GORM usando essa conexão já aberta
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		// PrepareStmt: false é OBRIGATÓRIO para o Transaction Mode (Porta 6543) do Supabase
		PrepareStmt: false,
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco de dados: %w", err)
	}

	return db, nil
}

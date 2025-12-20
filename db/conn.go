package db

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/guiezz/dashboard-api/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	// 1. Parse a configuração padrão do driver baseada na string de conexão
	dbConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao analisar configuração do banco: %w", err)
	}

	// 2. RESOLUÇÃO MANUAL DE DNS PARA IPV4
	// O problema: O Render só fala IPv4, mas o Supabase entrega IPv6 por padrão no DNS.
	// O driver pgx pega o IPv6 e trava. Aqui nós forçamos a pegar o IPv4.

	originalHost := dbConfig.Host

	// Busca todos os IPs desse host
	ips, err := net.LookupIP(originalHost)
	if err != nil {
		return nil, fmt.Errorf("erro de DNS ao buscar IP para %s: %w", originalHost, err)
	}

	var ipv4 string
	for _, ip := range ips {
		// Filtra apenas o que for IPv4
		if ip.To4() != nil {
			ipv4 = ip.String()
			break
		}
	}

	if ipv4 == "" {
		return nil, fmt.Errorf("nenhum endereço IPv4 encontrado para %s (O Render exige IPv4)", originalHost)
	}

	// Substitui o hostname pelo IP numérico na configuração
	dbConfig.Host = ipv4

	// 3. AJUSTE DE SSL/TLS (Crítico quando se usa IP direto)
	// Como trocamos o host pelo IP, o certificado SSL vai reclamar se não avisarmos
	// qual é o "nome original" do servidor (ServerName / SNI).
	if dbConfig.TLSConfig == nil {
		dbConfig.TLSConfig = &tls.Config{}
	}
	dbConfig.TLSConfig.ServerName = originalHost
	dbConfig.TLSConfig.MinVersion = tls.VersionTLS12

	// 4. Abre a conexão usando o stdlib do pgx
	sqlDB := stdlib.OpenDB(*dbConfig)

	// 5. Inicializa o GORM
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		// Desativa Prepared Statements (obrigatório para Supabase Transaction Mode/Porta 6543)
		PrepareStmt: false,
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco de dados: %w", err)
	}

	return db, nil
}

package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config armazena todas as variáveis de ambiente necessárias
type Config struct {
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string
	AppPort     string
	DBSSLMode   string
	JWTSecret   string
	FuncemeAPIURL string
	FrontendURLs  []string
}

// LoadConfig carrega as configurações do arquivo .env ou do ambiente do sistema.
func LoadConfig() *Config {
	// Tenta carregar o arquivo .env na raiz do projeto
	err := godotenv.Load()
	if err != nil {
		log.Println("Não foi possível encontrar o arquivo .env. As variáveis de ambiente devem ser definidas no sistema.")
	}

	// Retorna a struct Config preenchida com as variáveis de ambiente,
	// ou com valores padrão (fallback) se não estiverem definidas.
	return &Config{
		DBHost:     getEnv("POSTGRES_HOST", "localhost"),
		DBUser:     getEnv("POSTGRES_USER", "postgres"),
		DBPassword: getEnv("POSTGRES_PASSWORD", "postgres"),
		DBName:     getEnv("POSTGRES_DB", "postgres"),
		DBPort:     getEnv("POSTGRES_PORT", "5432"),
		AppPort:    getEnv("PORT", "8000"),
		DBSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		JWTSecret:  getEnv("JWT_SECRET", ""),
		FuncemeAPIURL: getEnv("FUNCEME_API_URL", "https://apil5.funceme.br/rpc/v1/reservatorio-series"),
		FrontendURLs:  parseOrigins(getEnv("FRONTEND_URL", "http://localhost:3000")),
	}
}

// GetDSN gera a string de conexão DSN (Data Source Name) para o GORM.
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode) // <--- Agora com as 6 variáveis
}

// getEnv é uma função auxiliar para obter uma variável de ambiente com um valor padrão.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.TrimSpace(value)
	}
	return fallback
}

// parseOrigins divide um string separada por vírgulas em um slice de origens CORS.
func parseOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

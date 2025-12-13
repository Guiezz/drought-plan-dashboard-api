package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config armazena todas as variáveis de ambiente necessárias
type Config struct {
	DBHost     string
	DBUser     string
	DBPassword string
	DBName     string
	DBPort     string
	AppPort    string
	// Se você tiver uma URL para a API Funceme, adicione aqui:
	// FuncemeAPIURL string
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
		// FuncemeAPIURL: getEnv("FUNCEME_API_URL", "http://api.funceme.br/v1/"),
	}
}

// GetDSN gera a string de conexão DSN (Data Source Name) para o GORM.
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort)
}

// getEnv é uma função auxiliar para obter uma variável de ambiente com um valor padrão.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

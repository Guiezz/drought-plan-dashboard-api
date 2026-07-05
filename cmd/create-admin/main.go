package main

import (
	"fmt"
	"log"
	"os"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/model"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@cogerh.gov.br"
	}

	senha := os.Getenv("ADMIN_PASSWORD")
	if senha == "" {
		senha = "cogerh123"
	}

	cfg := config.LoadConfig()
	dbConnection, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Erro ao gerar hash: %v", err)
	}

	admin := model.Usuario{
		Nome:      "Administrador Cogerh",
		Email:     email,
		SenhaHash: string(hash),
		Role:      "admin_cogerh",
	}

	result := dbConnection.Where(model.Usuario{Email: admin.Email}).FirstOrCreate(&admin)
	if result.Error != nil {
		log.Fatalf("Erro ao criar admin: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		fmt.Printf("Usuário %s já existe. Atualizando senha...\n", email)
		dbConnection.Model(&admin).Update("senha_hash", string(hash))
		fmt.Println("Senha atualizada com sucesso!")
	} else {
		fmt.Println("Administrador criado com sucesso!")
	}

	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Senha: %s\n", senha)
}

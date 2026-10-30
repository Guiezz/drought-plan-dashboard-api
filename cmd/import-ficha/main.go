// Comando import-ficha: grava a ficha técnica dos reservatórios a partir do CSV.
// Só atualiza os campos da ficha (por nome); não apaga nem recria dados.
//
//	go run ./cmd/import-ficha [-arquivo dados_importacao/ficha_tecnica.csv] [-dry-run]
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/ficha"
	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

var errDryRun = errors.New("dry-run: alterações desfeitas")

func main() {
	arquivo := flag.String("arquivo", "dados_importacao/ficha_tecnica.csv", "caminho do CSV da ficha técnica")
	dryRun := flag.Bool("dry-run", false, "mostra o resultado sem gravar no banco")
	flag.Parse()

	f, err := os.Open(*arquivo)
	if err != nil {
		log.Fatalf("Erro ao abrir %s: %v", *arquivo, err)
	}
	defer f.Close()

	regs, err := ficha.Parse(f)
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.ConnectDB(config.LoadConfig())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	// Só adiciona as colunas novas; não altera dados existentes.
	if err := database.AutoMigrate(&model.Reservatorio{}); err != nil {
		log.Fatalf("Erro ao migrar a tabela reservatorio: %v", err)
	}

	var atualizados int
	var desconhecidos []string
	err = database.Transaction(func(tx *gorm.DB) error {
		atualizados, desconhecidos, err = ficha.Aplicar(tx, regs)
		if err != nil {
			return err
		}
		if *dryRun {
			return errDryRun
		}
		return nil
	})
	if err != nil && !errors.Is(err, errDryRun) {
		log.Fatalf("Erro ao gravar a ficha: %v", err)
	}

	modo := "gravados"
	if *dryRun {
		modo = "simulados (nada foi gravado)"
	}
	fmt.Printf("Registros no CSV: %d · atualizados: %d %s\n", len(regs), atualizados, modo)
	if len(desconhecidos) > 0 {
		fmt.Printf("Nomes sem reservatório correspondente: %s\n", strings.Join(desconhecidos, ", "))
		os.Exit(2)
	}
}

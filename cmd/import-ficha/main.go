// Comando import-ficha: grava a ficha técnica dos reservatórios a partir do CSV.
// Só atualiza os campos da ficha (por nome); não apaga nem recria dados.
//
//	go run ./cmd/import-ficha [-arquivo dados_importacao/ficha_tecnica.csv] [-dry-run]
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/ficha"
)

func main() {
	arquivo := flag.String("arquivo", "dados_importacao/ficha_tecnica.csv", "caminho do CSV da ficha técnica")
	dryRun := flag.Bool("dry-run", false, "simula a importação numa transação desfeita no fim (não altera dados nem schema)")
	flag.Parse()

	f, err := ficha.AbrirCSV(*arquivo)
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

	// Cria as colunas da ficha (se faltarem) e grava numa só transação;
	// em -dry-run tudo é desfeito, inclusive as colunas.
	atualizados, desconhecidos, err := ficha.Executar(database, regs, *dryRun)
	if err != nil {
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

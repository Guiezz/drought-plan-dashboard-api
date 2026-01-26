package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func main() {
	// 1. Conexão
	cfg := config.LoadConfig()
	dbConn, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}

	// 2. Cria as tabelas se não existirem
	log.Println("🛠️  Verificando schema...")
	err = dbConn.AutoMigrate(
		&simulador.SimAcude{},
		&simulador.SimCAV{},
		&simulador.SimEvaporacao{},
		&simulador.SimVazao{},
	)
	if err != nil {
		log.Fatalf("Erro no migration: %v", err)
	}

	// 3. LIMPEZA: Apaga dados antigos para evitar duplicação e erros de FK
	log.Println("🧹 Limpando dados antigos da simulação...")
	// Ordem importa por causa das chaves estrangeiras
	dbConn.Exec("TRUNCATE TABLE sim_vazoes, sim_cav, sim_evaporacao, sim_acudes RESTART IDENTITY CASCADE")

	// 4. Abre Excel
	arquivoExcel := "testeAcudes.xlsx"
	f, err := excelize.OpenFile(arquivoExcel)
	if err != nil {
		log.Fatalf("❌ Erro ao abrir '%s': %v", arquivoExcel, err)
	}
	defer f.Close()

	// 5. Executa Importações na ordem correta
	importarAcudes(dbConn, f) // Primeiro Açudes (Pai)

	// Cria mapa de IDs válidos para filtrar órfãos
	validIDs := getAcudesMap(dbConn)
	log.Printf("ℹ️  Açudes válidos carregados: %d", len(validIDs))

	importarEvaporacao(dbConn, f)
	importarVazoes(dbConn, f, validIDs)
	importarCAV(dbConn, f, validIDs) // Passamos o mapa para validar

	log.Println("✅ Importação da Simulação concluída com sucesso!")
}

// --- Funções Auxiliares ---

// Carrega todos os IDs de açudes que realmente existem no banco
func getAcudesMap(db *gorm.DB) map[int]bool {
	var ids []int
	db.Model(&simulador.SimAcude{}).Pluck("codigo", &ids)
	validMap := make(map[int]bool)
	for _, id := range ids {
		validMap[id] = true
	}
	return validMap
}

func importarAcudes(db *gorm.DB, f *excelize.File) {
	nomeAba := encontrarAba(f, "acudes")
	if nomeAba == "" {
		log.Println("⚠️ Aba 'acudes' não encontrada.")
		return
	}
	log.Printf("💧 Importando Açudes...")
	rows, _ := f.GetRows(nomeAba)

	count := 0
	for i, row := range rows {
		if i == 0 || len(row) < 6 {
			continue
		}

		cod := parseInt(row[1])
		if cod == 0 {
			continue
		}

		db.Create(&simulador.SimAcude{
			Codigo:        cod,
			Nome:          row[0],
			Capacidade:    parseFloat(row[2]),
			Municipio:     row[3],
			EstacaoEvapID: parseInt(row[5]),
		})
		count++
	}
	log.Printf("✅ %d açudes importados.", count)
}

func importarCAV(db *gorm.DB, f *excelize.File, validAcudes map[int]bool) {
	nomeAba := encontrarAba(f, "cav")
	if nomeAba == "" {
		return
	}

	log.Printf("📊 Importando CAV...")
	rows, _ := f.GetRows(nomeAba)

	var lote []simulador.SimCAV
	ignored := 0
	total := 0

	for i, row := range rows {
		if i == 0 || len(row) < 4 {
			continue
		}

		cod := parseInt(row[0])

		// FILTRO DE SEGURANÇA: Se o açude não existe, pula
		if !validAcudes[cod] {
			ignored++
			continue
		}

		lote = append(lote, simulador.SimCAV{
			AcudeID: cod,
			Cota:    parseFloat(row[1]),
			Area:    parseFloat(row[2]),
			Volume:  parseFloat(row[3]),
		})
		total++

		if len(lote) >= 2000 {
			if err := db.CreateInBatches(lote, 2000).Error; err != nil {
				log.Printf("❌ Erro no lote CAV: %v", err)
			}
			lote = nil
			fmt.Printf("\r   -> Processados: %d", total)
		}
	}
	if len(lote) > 0 {
		db.CreateInBatches(lote, 2000)
	}
	fmt.Printf("\n✅ CAV Finalizado. Importados: %d | Ignorados (sem açude): %d\n", total, ignored)
}

func importarVazoes(db *gorm.DB, f *excelize.File, validAcudes map[int]bool) {
	nomeAba := encontrarAba(f, "vazoes")
	if nomeAba == "" {
		return
	}

	log.Printf("🌊 Importando Vazões...")
	rows, _ := f.GetRows(nomeAba)
	var lote []simulador.SimVazao

	for i, row := range rows {
		if i == 0 || len(row) < 14 {
			continue
		}

		cod := parseInt(row[0])
		if !validAcudes[cod] {
			continue
		}

		lote = append(lote, simulador.SimVazao{
			AcudeID: cod,
			Ano:     parseInt(row[1]),
			Jan:     parseFloat(row[2]), Fev: parseFloat(row[3]), Mar: parseFloat(row[4]),
			Abr: parseFloat(row[5]), Mai: parseFloat(row[6]), Jun: parseFloat(row[7]),
			Jul: parseFloat(row[8]), Ago: parseFloat(row[9]), Set: parseFloat(row[10]),
			Out: parseFloat(row[11]), Nov: parseFloat(row[12]), Dez: parseFloat(row[13]),
		})

		if len(lote) >= 2000 {
			db.CreateInBatches(lote, 2000)
			lote = nil
			fmt.Printf("\r   -> Vazões: %d...", i)
		}
	}
	if len(lote) > 0 {
		db.CreateInBatches(lote, 2000)
	}
	fmt.Println("\n✅ Vazões importadas.")
}

func importarEvaporacao(db *gorm.DB, f *excelize.File) {
	nomeAba := encontrarAba(f, "evaporacao")
	if nomeAba == "" {
		return
	}

	log.Printf("☀️  Importando Evaporação...")
	rows, _ := f.GetRows(nomeAba)
	for i, row := range rows {
		if i == 0 || len(row) < 15 {
			continue
		}
		db.Save(&simulador.SimEvaporacao{
			Codigo:    parseInt(row[1]),
			Municipio: row[0],
			Jan:       parseFloat(row[3]), Fev: parseFloat(row[4]), Mar: parseFloat(row[5]),
			Abr: parseFloat(row[6]), Mai: parseFloat(row[7]), Jun: parseFloat(row[8]),
			Jul: parseFloat(row[9]), Ago: parseFloat(row[10]), Set: parseFloat(row[11]),
			Out: parseFloat(row[12]), Nov: parseFloat(row[13]), Dez: parseFloat(row[14]),
		})
	}
	log.Println("✅ Evaporação importada.")
}

// --- Helpers ---
func encontrarAba(f *excelize.File, nomeAlvo string) string {
	for _, sheet := range f.GetSheetList() {
		if strings.EqualFold(strings.TrimSpace(sheet), nomeAlvo) {
			return sheet
		}
	}
	return ""
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0
	}
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return v
}

func parseInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

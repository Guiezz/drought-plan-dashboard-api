package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/model"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// 1. Carregar Configurações e Banco
	conf := config.LoadConfig()
	database, err := db.ConnectDB(conf)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}
	fmt.Println("✅ Conexão com o banco de dados realizada com sucesso!")

	// 2. LIMPEZA TOTAL (Se as tabelas existirem)
	limparBanco(database)

	// 3. MIGRAR BANCO (Cria as tabelas se não existirem)
	// Isso é essencial para não dar erro na hora de inserir os dados
	fmt.Println(">>> Verificando e criando schema do banco...")
	database.AutoMigrate(
		&model.Reservatorio{},
		&model.Monitoramento{},
		&model.UsoAgua{},
		&model.BalancoMensal{},
		&model.ComposicaoDemanda{},
		&model.OfertaDemanda{},
		&model.PlanoAcao{},
		&model.VolumeMeta{},
		&model.Responsavel{},
	)
	fmt.Println("✅ Schema do banco atualizado!")
	criarUsuarioAdmin(database)
	fmt.Println("------------------------------------------------")

	// 4. Caminho da pasta raiz e Processamento
	rootPath := "./dados_importacao"

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		log.Fatalf("Erro ao ler diretório raiz '%s': %v", rootPath, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf(">>> Processando reservatório: %s\n", entry.Name())
			folderPath := filepath.Join(rootPath, entry.Name())
			processarReservatorio(database, folderPath)
		}
	}
}

func criarUsuarioAdmin(db *gorm.DB) {
	fmt.Println(">>> Verificando usuário administrador...")
	senhaPlane := "cogerh123" // Senha padrão para testes

	// Gera o hash da senha
	hash, err := bcrypt.GenerateFromPassword([]byte(senhaPlane), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Erro ao gerar hash da senha: %v", err)
	}

	admin := model.Usuario{
		Nome:      "Administrador Cogerh",
		Email:     "admin@cogerh.gov.br",
		SenhaHash: string(hash),
		Role:      "admin_cogerh",
	}

	// FirstOrCreate verifica se o email já existe, se não, ele cria.
	if result := db.Where(model.Usuario{Email: admin.Email}).FirstOrCreate(&admin); result.Error != nil {
		log.Printf("⚠️ Erro ao criar usuário admin: %v", result.Error)
	} else {
		fmt.Printf("✅ Usuário admin pronto! Email: %s | Senha: %s\n", admin.Email, senhaPlane)
	}
}

func limparBanco(db *gorm.DB) {
	fmt.Println("!!! VERIFICANDO NECESSIDADE DE LIMPEZA !!!")

	// Verifica se a tabela principal existe antes de tentar limpar
	if !db.Migrator().HasTable(&model.Reservatorio{}) {
		fmt.Println("⚠️  Tabelas não encontradas. Pulando limpeza (provavelmente é a primeira execução).")
		return
	}

	fmt.Println("!!! INICIANDO LIMPEZA TOTAL E RESET DE IDs !!!")

	// O comando TRUNCATE limpa os dados e RESTART IDENTITY reseta os IDs para 1
	err := db.Exec(`
		TRUNCATE TABLE
			usuarios,
			historico_acoes,
			reservatorio,
			monitoramento,
			uso_agua,
			balanco_mensal,
			composicao_demanda,
			oferta_demanda,
			plano_acao,
			volume_meta,
			responsaveis
		RESTART IDENTITY CASCADE;
	`).Error

	if err != nil {
		// Se der erro aqui, apenas logamos, mas não matamos o processo, pois pode ser algo menor
		log.Printf("⚠️  Erro ao tentar limpar banco (pode ser ignorado se as tabelas estiverem vazias): %v", err)
	} else {
		fmt.Println("!!! BANCO LIMPO E IDs RESETADOS COM SUCESSO !!!")
	}
}

func processarReservatorio(db *gorm.DB, folderPath string) {
	// --- A. Processar Identificação ---
	var reservatorio model.Reservatorio
	var nome, municipio, descricao, nomeImg, nomeImgUsos, codigo string
	var lat, long, capacidade float64

	readExcel(folderPath, "identificacao", "", func(row []string) {
		if len(row) < 7 {
			return
		}
		lat, _ = parseFloat(row[1])
		long, _ = parseFloat(row[2])
		nome = row[3]
		municipio = row[4]
		descricao = row[0]
		nomeImg = row[5]
		nomeImgUsos = row[6]
		capacidade, _ = parseFloat(row[7])
		codigo = row[8]
	})

	if nome == "" {
		log.Printf("   [AVISO] Identificação não encontrada em %s (Pulando reservatório)", folderPath)
		return
	}

	resValues := model.Reservatorio{
		Nome:           nome,
		Municipio:      municipio,
		Descricao:      descricao,
		Lat:            lat,
		Long:           long,
		NomeImagem:     nomeImg,
		NomeImagemUsos: nomeImgUsos,
		Capacidadehm3:  capacidade,
		CodigoFunceme:  codigo,
	}

	if result := db.Where(model.Reservatorio{Nome: nome}).Attrs(resValues).FirstOrCreate(&reservatorio); result.Error != nil {
		log.Printf("   [ERRO] Falha ao criar reservatório %s: %v", nome, result.Error)
		return
	}
	fmt.Printf("   -> ID: %d (%s)\n", reservatorio.ID, reservatorio.Nome)

	// --- B. Monitoramento Histórico ---
	var monitoramentos []model.Monitoramento
	readExcel(folderPath, "monitoramento", "", func(row []string) {
		if len(row) < 3 || strings.Contains(strings.ToLower(row[0]), "data") {
			return
		}

		data, err := parseDate(row[0])
		if err != nil {
			return
		}

		vol, _ := parseFloat(row[1])
		volPorc, _ := parseFloat(row[2])

		monitoramentos = append(monitoramentos, model.Monitoramento{
			ReservatorioID:   reservatorio.ID,
			Data:             data,
			VolumeHm3:        vol,
			VolumePercentual: volPorc,
		})
	})
	if len(monitoramentos) > 0 {
		db.CreateInBatches(monitoramentos, 1000)
		fmt.Printf("      - %d monitoramentos importados.\n", len(monitoramentos))
	}

	// --- C. Composição Demanda ---
	readExcel(folderPath, "balanco_hidrico", "Composi", func(row []string) {
		if len(row) < 2 || strings.ToLower(row[0]) == "uso" {
			return
		}

		vazao, _ := parseFloat(row[1])
		db.Create(&model.ComposicaoDemanda{
			ReservatorioID: reservatorio.ID,
			Usos:           row[0],
			DemandasHm3:    vazao,
		})
	})

	// --- D. Oferta vs Demanda ---
	readExcel(folderPath, "balanco_hidrico", "Oferta", func(row []string) {
		if len(row) < 3 || strings.Contains(strings.ToLower(row[0]), "cenário") {
			return
		}

		oferta, _ := parseFloat(row[1])
		demanda, _ := parseFloat(row[2])

		db.Create(&model.OfertaDemanda{
			ReservatorioID: reservatorio.ID,
			Cenarios:       row[0],
			OfertaLs:       oferta,
			DemandaLs:      demanda,
		})
	})

	// --- E. Usos da Água ---
	readExcel(folderPath, "usos_agua", "", func(row []string) {
		if len(row) < 3 || strings.ToLower(row[0]) == "uso" {
			return
		}

		vazNormal, _ := parseFloat(row[1])
		vazEscassez, _ := parseFloat(row[2])

		db.Create(&model.UsoAgua{
			ReservatorioID: reservatorio.ID,
			Uso:            row[0],
			VazaoNormal:    vazNormal,
			VazaoEscassez:  vazEscassez,
		})
	})

	// --- F. Balanço Hídrico ---
	readExcel(folderPath, "balanco_hidrico", "Balan", func(row []string) {
		if len(row) < 4 || strings.Contains(row[0], "Mês") {
			return
		}

		afluencia, _ := parseFloat(row[1])
		evaporacao, _ := parseFloat(row[2])
		demanda, _ := parseFloat(row[3])

		db.Create(&model.BalancoMensal{
			ReservatorioID: reservatorio.ID,
			Mes:            row[0],
			AfluenciaM3s:   afluencia,
			EvaporacaoM3s:  evaporacao,
			DemandasM3s:    demanda,
		})
	})

	// --- G. Volume Meta ---
	readExcel(folderPath, "volume_meta", "", func(row []string) {
		if len(row) < 4 || strings.Contains(row[0], "Mes_Num") {
			return
		}

		// parseFloat corrigido já lida com % e divide por 100 se necessário
		meta1, _ := parseFloat(row[1])
		meta2, _ := parseFloat(row[2])
		meta3, _ := parseFloat(row[3])

		mesNome := row[0]
		mesNum := obterNumeroMes(mesNome)

		db.Create(&model.VolumeMeta{
			ReservatorioID: reservatorio.ID,
			MesNome:        mesNome,
			MesNum:         mesNum,
			Meta1v:         meta1,
			Meta2v:         meta2,
			Meta3v:         meta3,
		})
	})

	// --- H. Plano de Ação ---
	readExcel(folderPath, "plano_acao", "", func(row []string) {
		if len(row) < 8 || strings.Contains(row[0], "ESTADO") {
			return
		}

		db.Create(&model.PlanoAcao{
			ReservatorioID: reservatorio.ID,
			EstadoSeca:     row[0],
			TiposImpactos:  row[1],
			Problemas:      row[2],
			Acoes:          row[3],
			DescricaoAcao:  row[4],
			Responsaveis:   row[5],
			ClassesAcao:    row[6],
			Situacao:       row[7],
		})
	})

	// --- I. Responsáveis ---
	readExcel(folderPath, "responsaveis", "", func(row []string) {
		if len(row) == 0 || strings.Contains(strings.ToLower(row[0]), "grupo") {
			return
		}

		getCol := func(idx int) string {
			if idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		db.Create(&model.Responsavel{
			ReservatorioID: reservatorio.ID,
			Grupo:          getCol(0),
			Organizacao:    getCol(1),
			Setor:          getCol(2),
			Nome:           getCol(3),
			Cargo:          getCol(4),
		})
	})
}

// --- Funções Auxiliares ---

func readExcel(folder string, fileNamePart string, sheetNamePart string, processRow func([]string)) {
	files, _ := os.ReadDir(folder)
	var targetFile string

	for _, f := range files {
		if strings.Contains(strings.ToLower(f.Name()), strings.ToLower(fileNamePart)) && strings.HasSuffix(strings.ToLower(f.Name()), ".xlsx") {
			targetFile = filepath.Join(folder, f.Name())
			break
		}
	}

	if targetFile == "" {
		return
	}

	f, err := excelize.OpenFile(targetFile)
	if err != nil {
		log.Printf("   [ERRO] Falha ao abrir Excel %s: %v", filepath.Base(targetFile), err)
		return
	}
	defer f.Close()

	targetSheet := f.GetSheetName(0)

	if sheetNamePart != "" {
		found := false
		sheets := f.GetSheetList()
		for _, sheet := range sheets {
			if strings.Contains(strings.ToLower(sheet), strings.ToLower(sheetNamePart)) {
				targetSheet = sheet
				found = true
				break
			}
		}
		if !found {
			log.Printf("   [AVISO] Aba contendo '%s' não encontrada no arquivo %s.", sheetNamePart, filepath.Base(targetFile))
			return
		}
	}

	rows, err := f.GetRows(targetSheet)
	if err != nil {
		log.Printf("   [ERRO] Falha ao ler linhas da aba %s: %v", targetSheet, err)
		return
	}

	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		processRow(row)
	}
}

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	isPercent := strings.Contains(s, "%")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, ",", ".")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	if isPercent {
		val = val / 100.0
	}
	return val, nil
}

func parseDate(s string) (time.Time, error) {
	layouts := []string{"2006-01-02", "02/01/2006", "01-02-06", "1-2-06"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("data inválida")
}

func obterNumeroMes(mes string) int {
	mes = strings.TrimSpace(strings.ToLower(mes))
	prefixo := mes
	if len(mes) > 3 {
		prefixo = mes[:3]
	}
	switch prefixo {
	case "jan":
		return 1
	case "fev":
		return 2
	case "mar":
		return 3
	case "abr":
		return 4
	case "mai":
		return 5
	case "jun":
		return 6
	case "jul":
		return 7
	case "ago":
		return 8
	case "set":
		return 9
	case "out":
		return 10
	case "nov":
		return 11
	case "dez":
		return 12
	default:
		return 0
	}
}

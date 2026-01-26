package main

import (
	"log"
	"os"
	"strings"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/router"
	"github.com/guiezz/dashboard-api/usecase"

	// Importa o pacote docs gerado pelo Swaggo
	"github.com/guiezz/dashboard-api/docs"
)

// @title           API do dashboard de apoio a decisão dos planos de secas do estado do Ceará
// @version         1.0
// @description     API para gerenciamento e monitoramento de reservatórios hídricos do Ceará.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Suporte
// @contact.email   guilhermebessanojosaaraujo@gmail.com

// @host            localhost:8000
// @BasePath        /api
func main() {
	cfg := config.LoadConfig()

	// --- CORREÇÃO DO SWAGGER PARA O RENDER ---
	// Pega a URL externa configurada no ambiente
	externalURL := os.Getenv("RENDER_EXTERNAL_URL")
	if externalURL != "" {
		// O Swagger espera apenas o host (ex: meusite.onrender.com), sem https://
		host := strings.Replace(externalURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
		// Remove barra final se existir
		host = strings.TrimSuffix(host, "/")

		docs.SwaggerInfo.Host = host
		docs.SwaggerInfo.Schemes = []string{"https"}
		docs.SwaggerInfo.Description += "\n\n**Ambiente de Produção (Render)**"
	} else {
		docs.SwaggerInfo.Host = "localhost:" + cfg.AppPort
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	// -----------------------------------------

	// 2. Conexão DB
	dbConnection, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}

	// 2. Repositórios
	reservatorioRepo := repository.NewReservatorioRepository(dbConnection)
	planoAcaoRepo := repository.NewPlanoAcaoRepository(dbConnection)
	balancoRepo := repository.NewBalancoHidricoRepository(dbConnection)
	usoRepo := repository.NewUsoAguaRepository(dbConnection)
	respRepo := repository.NewResponsavelRepository(dbConnection)
	simulacaoRepo := repository.NewSimulacaoRepository(dbConnection)

	// 3. Serviços Internos (Cria as variáveis que estavam faltando)
	secaCalc := calculator.NewSecaCalculator()
	funcemeSvc := funceme.NewFuncemeService()
	simulacaoCalc := calculator.NewSimuladorHidrico()

	// 4. UseCases
	reservatorioUseCase := usecase.NewReservatorioUseCase(reservatorioRepo, planoAcaoRepo, secaCalc, funcemeSvc)
	planoAcaoUseCase := usecase.NewPlanoAcaoUseCase(planoAcaoRepo)
	balancoUseCase := usecase.NewBalancoHidricoUseCase(balancoRepo)
	usoUseCase := usecase.NewUsoAguaUseCase(usoRepo)
	respUseCase := usecase.NewResponsavelUseCase(respRepo)
	simulacaoUseCase := usecase.NewSimulacaoUseCase(simulacaoRepo, simulacaoCalc)

	// 5. Controllers
	reservatorioController := controller.NewReservatorioController(reservatorioUseCase)
	planoAcaoController := controller.NewPlanoAcaoController(planoAcaoUseCase)
	balancoController := controller.NewBalancoHidricoController(balancoUseCase)
	usoController := controller.NewUsoAguaController(usoUseCase)
	respController := controller.NewResponsavelController(respUseCase)
	simulacaoController := controller.NewSimulacaoController(simulacaoUseCase)

	// 6. Router
	server := router.SetupRouter(
		reservatorioController,
		planoAcaoController,
		balancoController,
		usoController,
		respController,
		simulacaoController,
	)

	server.Run(":" + cfg.AppPort)
}

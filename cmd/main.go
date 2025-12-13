package main

import (
	"log"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/router"
	"github.com/guiezz/dashboard-api/usecase"

	// IMPORTANTE: Isso será gerado no passo 1.4, mas precisamos importar
	// Troque "github.com/guiezz/dashboard-api" pelo nome exato do seu módulo no go.mod
	_ "github.com/guiezz/dashboard-api/docs"
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

	// 2. Conexão DB (Passando a config)
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

	// 3. Serviços Internos (Cria as variáveis que estavam faltando)
	secaCalc := calculator.NewSecaCalculator()
	funcemeSvc := funceme.NewFuncemeService()

	// 4. UseCases
	reservatorioUseCase := usecase.NewReservatorioUseCase(reservatorioRepo, planoAcaoRepo, secaCalc, funcemeSvc)
	planoAcaoUseCase := usecase.NewPlanoAcaoUseCase(planoAcaoRepo)
	balancoUseCase := usecase.NewBalancoHidricoUseCase(balancoRepo)
	usoUseCase := usecase.NewUsoAguaUseCase(usoRepo)
	respUseCase := usecase.NewResponsavelUseCase(respRepo)

	// 5. Controllers
	reservatorioController := controller.NewReservatorioController(reservatorioUseCase)
	planoAcaoController := controller.NewPlanoAcaoController(planoAcaoUseCase)
	balancoController := controller.NewBalancoHidricoController(balancoUseCase)
	usoController := controller.NewUsoAguaController(usoUseCase)
	respController := controller.NewResponsavelController(respUseCase)

	// 6. Router
	server := router.SetupRouter(
		reservatorioController,
		planoAcaoController,
		balancoController,
		usoController,
		respController,
	)

	server.Run(":" + cfg.AppPort)
}

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/guiezz/dashboard-api/config"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/internal/scheduler"
	"github.com/guiezz/dashboard-api/middleware"
	"github.com/guiezz/dashboard-api/model"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/router"
	"github.com/guiezz/dashboard-api/usecase"

	"github.com/guiezz/dashboard-api/docs"
)

// @title           API do dashboard de apoio a decisão dos planos de secas do estado do Ceará
// @version         1.0
// @description     API para gerenciamento e monitoramento de reservatórios hídricos do Ceará.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Suporte
// @contact.email   guilhermebessanojosaaraujo@gmail.com

// @host            localhost:8000
// @BasePath        /api
func main() {
	cfg := config.LoadConfig()

	// --- CORREÇÃO DO SWAGGER PARA O RENDER ---
	externalURL := os.Getenv("RENDER_EXTERNAL_URL")
	if externalURL != "" {
		host := strings.Replace(externalURL, "https://", "", 1)
		host = strings.Replace(host, "http://", "", 1)
		host = strings.TrimSuffix(host, "/")

		docs.SwaggerInfo.Host = host
		docs.SwaggerInfo.Schemes = []string{"https"}
		docs.SwaggerInfo.Description += "\n\n**Ambiente de Produção (Render)**"
	} else {
		docs.SwaggerInfo.Host = "localhost:" + cfg.AppPort
		docs.SwaggerInfo.Schemes = []string{"http"}
	}
	// -----------------------------------------

	// 1. Conexão DB
	dbConnection, err := db.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}

	if os.Getenv("APP_ENV") != "production" {
		dbConnection.AutoMigrate(&model.Usuario{}, &model.HistoricoAcao{})
	}

	// 2. Repositórios
	reservatorioRepo := repository.NewReservatorioRepository(dbConnection)
	planoAcaoRepo := repository.NewPlanoAcaoRepository(dbConnection)
	balancoRepo := repository.NewBalancoHidricoRepository(dbConnection)
	usoRepo := repository.NewUsoAguaRepository(dbConnection)
	respRepo := repository.NewResponsavelRepository(dbConnection)
	simulacaoRepo := repository.NewSimulacaoRepository(dbConnection)

	// 3. Serviços Internos
	secaCalc := calculator.NewSecaCalculator()
	funcemeSvc := funceme.NewFuncemeService(cfg.FuncemeAPIURL)
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

	// Criando a instância do AuthController e passando a conexão do banco + chave JWT
	authController := controller.NewAuthController(dbConnection, []byte(cfg.JWTSecret))

	// 6. Rate Limiters
	loginLimiter := middleware.NewRateLimiter(5, 1*time.Minute)
	simulacaoLimiter := middleware.NewRateLimiter(10, 1*time.Minute)

	// 7. Router
	server := router.SetupRouter(
		reservatorioController,
		planoAcaoController,
		balancoController,
		usoController,
		respController,
		simulacaoController,
		authController,
		cfg.FrontendURL,
		loginLimiter,
		simulacaoLimiter,
	)

	// 7. Scheduler automático da Funceme (executa a cada 24h)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.JWTSecret != "" {
		scheduler.Start(ctx, 24*time.Hour, 10*time.Second, 2*time.Second,
			reservatorioUseCase.ListReservoirIDs,
			reservatorioUseCase.AtualizarDadosFunceme,
		)
	}

	// 8. Servidor HTTP com graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: server,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Desligando servidor...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Erro ao desligar servidor: %v", err)
	}
}

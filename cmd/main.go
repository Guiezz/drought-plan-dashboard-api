package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/db"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/usecase"
)

func main() {
	// 1. Conexão via GORM
	dbConnection, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Erro ao conectar no banco: %v", err)
	}

	// 2. Inicialização das Camadas
	reservatorioRepo := repository.NewReservatorioRepository(dbConnection)

	secaCalc := calculator.NewSecaCalculator()
	funcemeSvc := funceme.NewFuncemeService()

	reservatorioUseCase := usecase.NewReservatorioUseCase(reservatorioRepo, secaCalc, funcemeSvc)
	// Controller
	reservatorioController := controller.NewReservatorioController(reservatorioUseCase)

	server := gin.Default()

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong!"})
	})

	server.Static("/static", "./static")

	api := server.Group("/api")
	{
		api.GET("/reservatorios", reservatorioController.GetReservatorios)

		res := api.Group("/reservatorios/:reservatorioId")
		{
			res.GET("/dashboard/summary", reservatorioController.GetDashboardSummary)
			res.GET("/identification", reservatorioController.GetIdentification)
			res.GET("/history", reservatorioController.GetHistory)

			res.GET("/ongoing-actions", reservatorioController.GetOngoingActions)
			res.GET("/completed-actions", reservatorioController.GetCompletedActions)
			res.GET("/action-plans", reservatorioController.GetActionPlans)
			res.GET("/action-plans/filters", reservatorioController.GetActionPlanFilters)
			res.GET("/usos-agua", reservatorioController.GetUsosAgua)
			res.GET("/responsaveis", reservatorioController.GetResponsaveis)

			res.GET("/chart/volume-data", reservatorioController.GetChartVolumeData)
			res.GET("/water-balance/static-charts", reservatorioController.GetWaterBalanceStaticCharts)

			res.POST("/update-funceme-data", reservatorioController.UpdateFuncemeData)
		}
	}

	server.Run(":8000")
}

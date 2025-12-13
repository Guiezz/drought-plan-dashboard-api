package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	resCtrl *controller.ReservatorioController,
	planoCtrl *controller.PlanoAcaoController,
	balancoCtrl *controller.BalancoHidricoController,
	usoCtrl *controller.UsoAguaController,
	respCtrl *controller.ResponsavelController,
) *gin.Engine {
	r := gin.Default()

	//Cors
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},

		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong!"})
	})

	r.Static("/static", "./static")

	api := r.Group("/api")
	{
		api.GET("/reservatorios", resCtrl.GetReservatorios)

		res := api.Group("/reservatorios/:reservatorioId")
		{
			// Dashboard e Identificação
			res.GET("/dashboard/summary", resCtrl.GetDashboardSummary)
			res.GET("/identification", resCtrl.GetIdentification)

			// Histórico e Gráficos
			res.GET("/history", resCtrl.GetHistory)
			res.GET("/dashboard/volume-chart", resCtrl.GetChartVolumeData)

			// Atualização Manual
			res.POST("/funceme-update", resCtrl.UpdateFuncemeData)

			// Planos de Ação
			res.GET("/ongoing-actions", planoCtrl.GetOngoingActions)
			res.GET("/completed-actions", planoCtrl.GetCompletedActions)
			res.GET("/action-plans", planoCtrl.GetActionPlans)
			res.GET("/action-plans/filters", planoCtrl.GetActionPlanFilters)

			// Balanço e Usos
			res.GET("/water-balance", balancoCtrl.GetCharts)
			res.GET("/water-uses", usoCtrl.GetUsos)
			res.GET("/responsibles", respCtrl.GetResponsaveis)
		}
	}

	return r
}

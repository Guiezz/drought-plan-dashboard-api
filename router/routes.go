package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/middleware"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	loginLimiter    = middleware.NewRateLimiter(5, 1*time.Minute)
	simulacaoLimiter = middleware.NewRateLimiter(10, 1*time.Minute)
)

func SetupRouter(
	resCtrl *controller.ReservatorioController,
	planoCtrl *controller.PlanoAcaoController,
	balancoCtrl *controller.BalancoHidricoController,
	usoCtrl *controller.UsoAguaController,
	respCtrl *controller.ResponsavelController,
	simCtrl *controller.SimulacaoController,
	authCtrl *controller.AuthController,
	frontendURL string,
) *gin.Engine {
	r := gin.Default()

	// Cors
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
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
		// Rota pública de login (com rate limit)
		api.POST("/login", loginLimiter.Middleware(), authCtrl.Login)

		api.GET("/reservatorios", resCtrl.GetReservatorios)

		res := api.Group("/reservatorios/:reservatorioId")
		{
			// Dashboard e Identificação (Públicas)
			res.GET("/dashboard/summary", resCtrl.GetDashboardSummary)
			res.GET("/identification", resCtrl.GetIdentification)

			// Histórico e Gráficos (Públicas)
			res.GET("/history", resCtrl.GetHistory)
			res.GET("/dashboard/volume-chart", resCtrl.GetChartVolumeData)

			// Gatilhos do PGPS
			res.GET("/gatilhos-pgps", resCtrl.GetGatilhosPGPS)

			// Planos de Ação (Leitura Pública)
			res.GET("/not-started-actions", planoCtrl.GetNotStartedActions)
			res.GET("/ongoing-actions", planoCtrl.GetOngoingActions)
			res.GET("/completed-actions", planoCtrl.GetCompletedActions)
			res.GET("/action-plans", planoCtrl.GetActionPlans)
			res.GET("/action-plans/filters", planoCtrl.GetActionPlanFilters)

			// Grupo Protegido para Edição de Planos de Ação
			protectedRes := res.Group("/")
			protectedRes.Use(middleware.AuthMiddleware(authCtrl.JwtSecret))
			{
				// Iremos criar este endpoint no PlanoAcaoController a seguir
				protectedRes.PUT("/action-plans/:acaoId/status", planoCtrl.UpdateStatus)
			}

			// Balanço e Usos (Públicas)
			res.GET("/water-balance", balancoCtrl.GetCharts)
			res.GET("/water-uses", usoCtrl.GetUsos)
			res.GET("/responsibles", respCtrl.GetResponsaveis)

			// Simulação (Públicas, com rate limit no /run)
			sim := api.Group("/simulacao")
			{
				sim.POST("/run", simulacaoLimiter.Middleware(), simCtrl.Simular)
				sim.GET("/acudes", simCtrl.ListarAcudes)
				sim.GET("/anos", simCtrl.ListarAnos)
			}
		}
	}

	return r
}

package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/docs"
	"github.com/guiezz/dashboard-api/middleware"
)

func SetupRouter(
	resCtrl *controller.ReservatorioController,
	planoCtrl *controller.PlanoAcaoController,
	balancoCtrl *controller.BalancoHidricoController,
	usoCtrl *controller.UsoAguaController,
	respCtrl *controller.ResponsavelController,
	simCtrl *controller.SimulacaoController,
	authCtrl *controller.AuthController,
	frontendURLs []string,
	loginLimiter, simulacaoLimiter *middleware.RateLimiter,
) *gin.Engine {
	r := gin.Default()

	// Cors
	r.Use(cors.New(cors.Config{
		AllowOrigins: frontendURLs,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong!"})
	})

	r.GET("/openapi.json", func(c *gin.Context) {
		c.Data(200, "application/json", []byte(docs.SwaggerInfo.ReadDoc()))
	})

	r.GET("/docs", func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, `<!DOCTYPE html>
<html>
<head>
  <title>API Dashboard - Documentação</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <div id="api-reference" data-url="/openapi.json"></div>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`)
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

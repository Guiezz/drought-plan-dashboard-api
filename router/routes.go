package router

import (
	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"
)

func SetupRouter(
	resCtrl *controller.ReservatorioController,
	planoCtrl *controller.PlanoAcaoController,
	balancoCtrl *controller.BalancoHidricoController,
) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong!"})
	})

	r.Static("/static", "./static")

	api := r.Group("/api")
	{
		api.GET("/reservatorios", resCtrl.GetReservatorios)

		res := api.Group("/reservatorios/:reservatorioId")
		{
			res.GET("/dashboard/summary", resCtrl.GetDashboardSummary)
			res.GET("/identification", resCtrl.GetIdentification)
			res.GET("/history", resCtrl.GetHistory)

			// ... Mova todas as outras rotas para cá ...

			// Rotas de Plano de Ação (Ainda usando o controller antigo por enquanto,
			// mas idealmente teria seu próprio controller)
			res.GET("/ongoing-actions", planoCtrl.GetOngoingActions)
			res.GET("/completed-actions", planoCtrl.GetCompletedActions)
			res.GET("/action-plans", planoCtrl.GetActionPlans)
			res.GET("/action-plans/filters", planoCtrl.GetActionPlanFilters)
			// ...
			//
			res.GET("/water-balance", balancoCtrl.GetCharts)
		}
	}

	return r
}

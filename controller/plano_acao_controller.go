package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
)

type PlanoAcaoController struct {
	useCase *usecase.PlanoAcaoUseCase
}

func NewPlanoAcaoController(useCase *usecase.PlanoAcaoUseCase) *PlanoAcaoController {
	return &PlanoAcaoController{
		useCase: useCase,
	}
}

func (c *PlanoAcaoController) GetOngoingActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	acoes, err := c.useCase.Listar(id, "Em andamento", "", "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

func (c *PlanoAcaoController) GetCompletedActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	acoes, err := c.useCase.Listar(id, "Concluído", "", "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

func (c *PlanoAcaoController) GetActionPlans(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	estado := ctx.Query("estado")
	impacto := ctx.Query("impacto")
	problema := ctx.Query("problema")
	acao := ctx.Query("acao")

	planos, err := c.useCase.Listar(id, "", estado, impacto, problema, acao)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, planos)
}

func (c *PlanoAcaoController) GetActionPlanFilters(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	filtros, err := c.useCase.ObterFiltros(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, filtros)
}

// Helper auxiliar
func (c *PlanoAcaoController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)
	return id
}

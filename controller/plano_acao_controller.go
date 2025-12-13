package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
	// Importe o model para o Swagger reconhecer os tipos de retorno
)

type PlanoAcaoController struct {
	useCase *usecase.PlanoAcaoUseCase
}

func NewPlanoAcaoController(useCase *usecase.PlanoAcaoUseCase) *PlanoAcaoController {
	return &PlanoAcaoController{useCase: useCase}
}

// GetOngoingActions godoc
// @Summary      Ações em Andamento
// @Description  Lista planos de ação com situação 'Em andamento'
// @Tags         Planos de Ação
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {array}   model.PlanoAcao
// @Router       /reservatorios/{reservatorioId}/ongoing-actions [get]
func (c *PlanoAcaoController) GetOngoingActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	acoes, err := c.useCase.Listar(id, "Em andamento", "", "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

// GetCompletedActions godoc
// @Summary      Ações Concluídas
// @Description  Lista planos de ação com situação 'Concluído'
// @Tags         Planos de Ação
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {array}   model.PlanoAcao
// @Router       /reservatorios/{reservatorioId}/completed-actions [get]
func (c *PlanoAcaoController) GetCompletedActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	acoes, err := c.useCase.Listar(id, "Concluído", "", "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

// GetActionPlans godoc
// @Summary      Listar Planos de Ação (Com Filtros)
// @Description  Busca planos baseados nos filtros opcionais (query params)
// @Tags         Planos de Ação
// @Param        reservatorioId   path      int     true  "ID do Reservatório"
// @Param        estado           query     string  false "Estado de Seca (ex: SECA, ALERTA)"
// @Param        impacto          query     string  false "Tipo de Impacto"
// @Param        problema         query     string  false "Problema Identificado"
// @Param        acao             query     string  false "Nome da Ação"
// @Produce      json
// @Success      200  {array}   model.PlanoAcao
// @Router       /reservatorios/{reservatorioId}/action-plans [get]
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

// GetActionPlanFilters godoc
// @Summary      Opções de Filtros
// @Description  Retorna as listas de valores únicos disponíveis para filtrar os planos
// @Tags         Planos de Ação
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.FiltrosPlanoAcao
// @Router       /reservatorios/{reservatorioId}/action-plans/filters [get]
func (c *PlanoAcaoController) GetActionPlanFilters(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	filtros, err := c.useCase.ObterFiltros(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, filtros)
}

func (c *PlanoAcaoController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)
	return id
}

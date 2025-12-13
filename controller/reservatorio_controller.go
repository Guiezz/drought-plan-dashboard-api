package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/model"
	"github.com/guiezz/dashboard-api/usecase"
)

type ReservatorioController struct {
	useCase *usecase.ReservatorioUseCase
}

func NewReservatorioController(useCase *usecase.ReservatorioUseCase) *ReservatorioController {
	return &ReservatorioController{
		useCase: useCase,
	}
}

// GetReservatorios godoc
// @Summary      Lista reservatórios
// @Description  Retorna a lista de reservatórios cadastrados para seleção
// @Tags         Reservatórios
// @Produce      json
// @Success      200  {array}   model.Reservatorio
// @Router       /reservatorios [get]
func (c *ReservatorioController) GetReservatorios(ctx *gin.Context) {
	reservatorios, err := c.useCase.ListarTodos()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Erro ao buscar reservatórios",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, reservatorios)
}

// GetDashboardSummary godoc
// @Summary      Resumo do Dashboard
// @Description  Retorna volume, status de seca e dias desde mudança de status
// @Tags         Dashboard
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.DashboardResumo
// @Router       /reservatorios/{reservatorioId}/dashboard/summary [get]
func (c *ReservatorioController) GetDashboardSummary(ctx *gin.Context) {
	idStr := ctx.Param("reservatorioId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "ID do reservatório deve ser um número inteiro",
		})
		return
	}

	resumo, err := c.useCase.ObterResumoDashboard(id)
	if err != nil {
		// Aqui poderíamos checar se o erro é "não encontrado" vs "erro de banco"
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Dados não encontrados para este reservatório",
		})
		return
	}

	ctx.JSON(http.StatusOK, resumo)
}

// --- NOVOS HANDLERS ---

func (c *ReservatorioController) GetUsosAgua(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	usos, err := c.useCase.ListarUsosAgua(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, usos)
}

func (c *ReservatorioController) GetResponsaveis(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	resps, err := c.useCase.ListarResponsaveis(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resps)
}

// Helper para pegar ID da URL
func (c *ReservatorioController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)
	return id
}

func (c *ReservatorioController) GetChartVolumeData(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	dados, err := c.useCase.ObterDadosGrafico(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dados)
}

// GET /api/reservatorios/{reservatorio_id}/identification
func (c *ReservatorioController) GetIdentification(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	detalhes, err := c.useCase.ObterDetalhesReservatorio(id)
	if err != nil {
		// Retorna 404 se não achar
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Reservatório não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, detalhes)
}

// GET /api/reservatorios/{reservatorio_id}/history
func (c *ReservatorioController) GetHistory(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	historico, err := c.useCase.ObterHistoricoTabular(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Se vazio, retornar array vazio [] e não null
	if historico == nil {
		historico = []model.HistoricoTabela{}
	}

	ctx.JSON(http.StatusOK, historico)
}

func (c *ReservatorioController) UpdateFuncemeData(ctx *gin.Context) {
	id := c.getIdParam(ctx)

	qtdNovos, err := c.useCase.AtualizarDadosFunceme(id)
	if err != nil {
		// Verifica se é erro de regra de negócio ou servidor
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if qtdNovos == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "Nenhum registro novo. O banco já está atualizado.",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": fmt.Sprintf("%d novos registros de monitoramento foram adicionados.", qtdNovos),
	})
}

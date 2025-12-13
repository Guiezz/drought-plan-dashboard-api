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
// @Description  Retorna a lista de todos os reservatórios cadastrados
// @Tags         Reservatórios
// @Produce      json
// @Success      200  {array}   model.Reservatorio
// @Router       /reservatorios [get]
func (c *ReservatorioController) GetReservatorios(ctx *gin.Context) {
	reservatorios, err := c.useCase.ListarTodos()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar reservatórios", "details": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reservatorios)
}

// GetDashboardSummary godoc
// @Summary      Resumo do Dashboard
// @Description  Retorna volume atual, status de seca e resumo de ações recomendadas
// @Tags         Dashboard
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.DashboardResumo
// @Router       /reservatorios/{reservatorioId}/dashboard/summary [get]
func (c *ReservatorioController) GetDashboardSummary(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	resumo, err := c.useCase.ObterResumoDashboard(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Dados não encontrados"})
		return
	}
	ctx.JSON(http.StatusOK, resumo)
}

// GetChartVolumeData godoc
// @Summary      Dados do Gráfico de Volume
// @Description  Retorna série histórica de volume e linhas de meta (V.O.)
// @Tags         Gráficos
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {array}   model.GraficoVolumeData
// @Router       /reservatorios/{reservatorioId}/dashboard/volume-chart [get]
func (c *ReservatorioController) GetChartVolumeData(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	dados, err := c.useCase.ObterDadosGrafico(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dados)
}

// GetIdentification godoc
// @Summary      Identificação e Localização
// @Description  Retorna dados cadastrais, URLs de imagens e localização
// @Tags         Reservatórios
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.ReservatorioDetalhes
// @Router       /reservatorios/{reservatorioId}/identification [get]
func (c *ReservatorioController) GetIdentification(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	detalhes, err := c.useCase.ObterDetalhesReservatorio(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Reservatório não encontrado"})
		return
	}
	ctx.JSON(http.StatusOK, detalhes)
}

// GetHistory godoc
// @Summary      Histórico Tabular
// @Description  Retorna tabela completa de monitoramento com status de seca calculado
// @Tags         Monitoramento
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {array}   model.HistoricoTabela
// @Router       /reservatorios/{reservatorioId}/history [get]
func (c *ReservatorioController) GetHistory(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	historico, err := c.useCase.ObterHistoricoTabular(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if historico == nil {
		historico = []model.HistoricoTabela{}
	}
	ctx.JSON(http.StatusOK, historico)
}

// UpdateFuncemeData godoc
// @Summary      Forçar Atualização Funceme
// @Description  Busca dados novos na API da Funceme e salva no banco se houver novidades
// @Tags         Admin
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /reservatorios/{reservatorioId}/funceme-update [post]
func (c *ReservatorioController) UpdateFuncemeData(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	qtdNovos, err := c.useCase.AtualizarDadosFunceme(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if qtdNovos == 0 {
		ctx.JSON(http.StatusOK, gin.H{"status": "O banco já está atualizado."})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": fmt.Sprintf("%d novos registros adicionados.", qtdNovos)})
}

// Helper interno (não exposto no Swagger)
func (c *ReservatorioController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)
	return id
}

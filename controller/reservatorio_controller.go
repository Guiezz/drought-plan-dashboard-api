package controller

import (
	"log"
	"net/http"
	"strconv"
	"time"

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
		log.Printf("[ERRO] GetReservatorios: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar reservatórios"})
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
		log.Printf("[ERRO] GetChartVolumeData: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar dados do gráfico"})
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
		log.Printf("[ERRO] GetHistory: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar histórico"})
		return
	}
	if historico == nil {
		historico = []model.HistoricoTabela{}
	}
	ctx.JSON(http.StatusOK, historico)
}

// GetGatilhosPGPS godoc
// @Summary      Gatilhos do PGPS
// @Description  Retorna os gatilhos mensais do PGPS (meta1v, meta2v, meta3v) em hm³ para cada mês
// @Tags         Reservatórios
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.GatilhosPGPSResponse
// @Router       /reservatorios/{reservatorioId}/gatilhos-pgps [get]
func (c *ReservatorioController) GetGatilhosPGPS(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	dados, err := c.useCase.ObterGatilhosPGPS(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Dados não encontrados"})
		return
	}
	ctx.JSON(http.StatusOK, dados)
}

type BackfillRequest struct {
	DataInicio string `json:"data_inicio" binding:"required"`
}

func (c *ReservatorioController) BackfillDados(ctx *gin.Context) {
	id := c.getIdParam(ctx)

	var req BackfillRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Campo 'data_inicio' é obrigatório (formato YYYY-MM-DD)"})
		return
	}

	if _, err := time.Parse("2006-01-02", req.DataInicio); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Formato de data inválido. Use YYYY-MM-DD"})
		return
	}

	role, _ := ctx.Get("role")
	log.Printf("[INFO] Backfill solicitado para reservatório %d desde %s pelo usuário com role: %v", id, req.DataInicio, role)

	novosRegistros, err := c.useCase.BackfillFunceme(uint(id), req.DataInicio)
	if err != nil {
		log.Printf("[ERRO] BackfillFunceme: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao realizar backfill"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Backfill concluído",
		"registros": novosRegistros,
	})
}

// Helper interno (não exposto no Swagger)
func (c *ReservatorioController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("[AVISO] getIdParam: reservatorioId inválido: %q", idStr)
		return 0
	}
	return id
}

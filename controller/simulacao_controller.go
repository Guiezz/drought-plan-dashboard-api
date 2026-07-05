package controller

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/guiezz/dashboard-api/usecase"
)

type SimulacaoController struct {
	useCase *usecase.SimulacaoUseCase
}

func NewSimulacaoController(useCase *usecase.SimulacaoUseCase) *SimulacaoController {
	return &SimulacaoController{useCase: useCase}
}

// @Summary      Executa a simulação hídrica
// @Description  Calcula o balanço hídrico mensal com base em dados históricos
// @Tags         Simulacao
// @Accept       json
// @Produce      json
// @Param        request body simulador.SimulacaoRequest true "Parâmetros da Simulação"
// @Success      200 {object} simulador.SimulacaoResponse
// @Router       /api/simulacao/run [post]
func (c *SimulacaoController) Simular(ctx *gin.Context) {
	var req simulador.SimulacaoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Cenarios) > 0 {
		// Modo multi-cenário
		resultado, err := c.useCase.ExecutarMultiCenario(req)
		if err != nil {
			log.Printf("[ERRO] Simular (multi-cenário): %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao executar simulação"})
			return
		}
		ctx.JSON(http.StatusOK, resultado)
		return
	}

	// Modo single (comportamento original)
	resultado, err := c.useCase.Executar(req)
	if err != nil {
		log.Printf("[ERRO] Simular: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao executar simulação"})
		return
	}

	ctx.JSON(http.StatusOK, resultado)
}

// @Summary      Lista açudes disponíveis para simulação
// @Tags         Simulacao
// @Success      200 {array} simulador.SimAcude
// @Router       /api/simulacao/acudes [get]
func (c *SimulacaoController) ListarAcudes(ctx *gin.Context) {
	acudes, err := c.useCase.ListarOpcoes()
	if err != nil {
		log.Printf("[ERRO] ListarAcudes: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar açudes"})
		return
	}
	ctx.JSON(http.StatusOK, acudes)
}

// @Summary      Lista anos com vazão cadastrada para um açude
// @Tags         Simulacao
// @Param        reservatorio_id query int true "ID do reservatório"
// @Success      200 {object} map[string][]int
// @Router       /api/simulacao/anos [get]
func (c *SimulacaoController) ListarAnos(ctx *gin.Context) {
	reservatorioIDStr := ctx.Query("reservatorio_id")
	if reservatorioIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "reservatorio_id é obrigatório"})
		return
	}

	reservatorioID, err := parseIntSafe(reservatorioIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "reservatorio_id deve ser um número inteiro"})
		return
	}

	anos, err := c.useCase.ListarAnos(reservatorioID)
	if err != nil {
		log.Printf("[ERRO] ListarAnos: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao listar anos"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"anos": anos})
}

func parseIntSafe(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("não é um número")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

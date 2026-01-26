package controller

import (
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

	resultado, err := c.useCase.Executar(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acudes)
}

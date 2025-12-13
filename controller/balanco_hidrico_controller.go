package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
)

type BalancoHidricoController struct {
	useCase *usecase.BalancoHidricoUseCase
}

func NewBalancoHidricoController(useCase *usecase.BalancoHidricoUseCase) *BalancoHidricoController {
	return &BalancoHidricoController{useCase: useCase}
}

// GetCharts godoc
// @Summary      Gráficos de Balanço Hídrico
// @Description  Retorna dados formatados para os 3 gráficos: Balanço Mensal, Demanda e Oferta
// @Tags         Gráficos
// @Param        reservatorioId   path      int  true  "ID do Reservatório"
// @Produce      json
// @Success      200  {object}  model.BalancoHidricoResumo
// @Router       /reservatorios/{reservatorioId}/water-balance [get]
func (c *BalancoHidricoController) GetCharts(ctx *gin.Context) {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)

	dados, err := c.useCase.ObterResumo(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dados)
}

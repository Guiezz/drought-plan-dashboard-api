package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
)

type UsoAguaController struct {
	useCase *usecase.UsoAguaUseCase
}

func NewUsoAguaController(useCase *usecase.UsoAguaUseCase) *UsoAguaController {
	return &UsoAguaController{useCase: useCase}
}

func (c *UsoAguaController) GetUsos(ctx *gin.Context) {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)

	usos, err := c.useCase.Listar(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, usos)
}

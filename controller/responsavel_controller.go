package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
)

type ResponsavelController struct {
	useCase *usecase.ResponsavelUseCase
}

func NewResponsavelController(useCase *usecase.ResponsavelUseCase) *ResponsavelController {
	return &ResponsavelController{useCase: useCase}
}

func (c *ResponsavelController) GetResponsaveis(ctx *gin.Context) {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)

	resps, err := c.useCase.Listar(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resps)
}

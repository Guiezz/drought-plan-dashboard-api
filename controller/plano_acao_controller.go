package controller

import (
	"log" // Novo import para logs
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/usecase"
)

type PlanoAcaoController struct {
	useCase *usecase.PlanoAcaoUseCase
}

func NewPlanoAcaoController(useCase *usecase.PlanoAcaoUseCase) *PlanoAcaoController {
	return &PlanoAcaoController{useCase: useCase}
}

// ... (GetNotStartedActions, GetOngoingActions, GetCompletedActions, GetActionPlans, GetActionPlanFilters se mantêm iguais ao seu código atual) ...
func (c *PlanoAcaoController) GetNotStartedActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	estado := ctx.Query("estado") // Captura o filtro do estado de seca
	acoes, err := c.useCase.Listar(id, "Não iniciado", estado, "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

func (c *PlanoAcaoController) GetOngoingActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	estado := ctx.Query("estado") // Captura o filtro do estado de seca
	acoes, err := c.useCase.Listar(id, "Em andamento", estado, "", "", "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, acoes)
}

func (c *PlanoAcaoController) GetCompletedActions(ctx *gin.Context) {
	id := c.getIdParam(ctx)
	estado := ctx.Query("estado") // Captura o filtro do estado de seca
	acoes, err := c.useCase.Listar(id, "Concluído", estado, "", "", "")
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

type UpdateStatusRequest struct {
	Situacao string `json:"situacao" binding:"required"`
}

func (c *PlanoAcaoController) UpdateStatus(ctx *gin.Context) {
	acaoIdStr := ctx.Param("acaoId")
	acaoId, err := strconv.ParseUint(acaoIdStr, 10, 32)
	if err != nil {
		log.Printf("[ERRO Controller] ID da ação inválido: %v", acaoIdStr)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ID da ação inválido"})
		return
	}

	var req UpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("[ERRO Controller] Erro no JSON do body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "O campo 'situacao' é obrigatório"})
		return
	}

	usuarioIdInterface, exists := ctx.Get("usuario_id")
	if !exists {
		log.Println("[ERRO Controller] Usuário não encontrado no contexto do Gin")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não identificado na sessão"})
		return
	}

	// CONVERSÃO SEGURA
	usuarioId, ok := usuarioIdInterface.(uint)
	if !ok {
		log.Println("[ERRO Controller] Falha ao converter usuario_id para uint")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Erro de tipagem no ID do usuário"})
		return
	}

	log.Printf("[INFO Controller] Solicitando update da Ação %d para status '%s' pelo Usuário %d", acaoId, req.Situacao, usuarioId)

	if err := c.useCase.AtualizarStatus(uint(acaoId), usuarioId, req.Situacao); err != nil {
		log.Printf("[ERRO Controller] UseCase retornou erro: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar o status: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status da ação atualizado com sucesso!"})
}

func (c *PlanoAcaoController) getIdParam(ctx *gin.Context) int {
	idStr := ctx.Param("reservatorioId")
	id, _ := strconv.Atoi(idStr)
	return id
}

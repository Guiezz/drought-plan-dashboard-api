package usecase

import (
	"errors"
	"testing"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
)

func TestPlanoAcaoListar(t *testing.T) {
	mockRepo := new(mockPlanoAcaoRepo)
	uc := NewPlanoAcaoUseCase(mockRepo)

	planos := []model.PlanoAcao{
		{Acoes: "Ação 1", Situacao: "Não iniciado"},
	}

	mockRepo.On("Listar", 1, "Não iniciado", "", "", "", "").Return(planos, nil).Once()

	resultado, err := uc.Listar(1, "Não iniciado", "", "", "", "")
	assert.NoError(t, err)
	assert.Len(t, resultado, 1)
	assert.Equal(t, "Ação 1", resultado[0].Acoes)
	mockRepo.AssertExpectations(t)
}

func TestPlanoAcaoListarErro(t *testing.T) {
	mockRepo := new(mockPlanoAcaoRepo)
	uc := NewPlanoAcaoUseCase(mockRepo)

	mockRepo.On("Listar", 1, "", "", "", "", "").Return([]model.PlanoAcao{}, errors.New("db error")).Once()

	resultado, err := uc.Listar(1, "", "", "", "", "")
	assert.Error(t, err)
	assert.Empty(t, resultado)
	mockRepo.AssertExpectations(t)
}

func TestPlanoAcaoAtualizarStatus(t *testing.T) {
	mockRepo := new(mockPlanoAcaoRepo)
	uc := NewPlanoAcaoUseCase(mockRepo)

	mockRepo.On("AtualizarStatus", uint(1), uint(1), "Em andamento").Return(nil).Once()

	err := uc.AtualizarStatus(1, 1, "Em andamento")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

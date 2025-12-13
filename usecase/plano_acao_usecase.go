package usecase

import (
	"github.com/guiezz/dashboard-api/model"
)

// Interface local para desacoplar do repositório concreto
type PlanoAcaoRepositoryInterface interface {
	Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error)
	ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error)
}

type PlanoAcaoUseCase struct {
	repo PlanoAcaoRepositoryInterface
}

func NewPlanoAcaoUseCase(repo PlanoAcaoRepositoryInterface) *PlanoAcaoUseCase {
	return &PlanoAcaoUseCase{repo: repo}
}

func (uc *PlanoAcaoUseCase) Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error) {
	// Aqui você poderia adicionar lógicas de negócio específicas se precisasse
	// Por exemplo: ordenar por gravidade, formatar textos, etc.
	return uc.repo.Listar(reservatorioID, situacao, estado, impacto, problema, acao)
}

func (uc *PlanoAcaoUseCase) ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error) {
	return uc.repo.ObterFiltros(reservatorioID)
}

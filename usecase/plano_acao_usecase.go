package usecase

import (
	"github.com/guiezz/dashboard-api/model"
)

type PlanoAcaoRepositoryInterface interface {
	Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error)
	ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error)
	AtualizarStatus(acaoID uint, usuarioID uint, novaSituacao string) error
}

type PlanoAcaoUseCase struct {
	repo PlanoAcaoRepositoryInterface
}

func NewPlanoAcaoUseCase(repo PlanoAcaoRepositoryInterface) *PlanoAcaoUseCase {
	return &PlanoAcaoUseCase{repo: repo}
}

func (uc *PlanoAcaoUseCase) Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error) {
	return uc.repo.Listar(reservatorioID, situacao, estado, impacto, problema, acao)
}

func (uc *PlanoAcaoUseCase) ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error) {
	return uc.repo.ObterFiltros(reservatorioID)
}

func (uc *PlanoAcaoUseCase) AtualizarStatus(acaoID uint, usuarioID uint, novaSituacao string) error {
	return uc.repo.AtualizarStatus(acaoID, usuarioID, novaSituacao)
}

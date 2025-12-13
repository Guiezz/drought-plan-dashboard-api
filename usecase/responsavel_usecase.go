package usecase

import "github.com/guiezz/dashboard-api/model"

type ResponsavelRepositoryInterface interface {
	Listar(reservatorioID int) ([]model.Responsavel, error)
}

type ResponsavelUseCase struct {
	repo ResponsavelRepositoryInterface
}

func NewResponsavelUseCase(repo ResponsavelRepositoryInterface) *ResponsavelUseCase {
	return &ResponsavelUseCase{repo: repo}
}

func (uc *ResponsavelUseCase) Listar(reservatorioID int) ([]model.Responsavel, error) {
	return uc.repo.Listar(reservatorioID)
}

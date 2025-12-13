package usecase

import "github.com/guiezz/dashboard-api/model"

type UsoAguaRepositoryInterface interface {
	Listar(reservatorioID int) ([]model.UsoAgua, error)
}

type UsoAguaUseCase struct {
	repo UsoAguaRepositoryInterface
}

func NewUsoAguaUseCase(repo UsoAguaRepositoryInterface) *UsoAguaUseCase {
	return &UsoAguaUseCase{repo: repo}
}

func (uc *UsoAguaUseCase) Listar(reservatorioID int) ([]model.UsoAgua, error) {
	return uc.repo.Listar(reservatorioID)
}

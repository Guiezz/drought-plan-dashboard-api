package usecase

import (
	"github.com/guiezz/dashboard-api/model"
)

type BalancoHidricoRepositoryInterface interface {
	GetBalancoMensal(reservatorioID int) ([]model.BalancoMensal, error)
	GetComposicaoDemanda(reservatorioID int) ([]model.ComposicaoDemanda, error)
	GetOfertaDemanda(reservatorioID int) ([]model.OfertaDemanda, error)
}

type BalancoHidricoUseCase struct {
	repo BalancoHidricoRepositoryInterface
}

func NewBalancoHidricoUseCase(repo BalancoHidricoRepositoryInterface) *BalancoHidricoUseCase {
	return &BalancoHidricoUseCase{repo: repo}
}

func (uc *BalancoHidricoUseCase) ObterResumo(reservatorioID int) (*model.BalancoHidricoResumo, error) {
	// 1. Busca dados
	bm, err := uc.repo.GetBalancoMensal(reservatorioID)
	if err != nil {
		return nil, err
	}

	cd, err := uc.repo.GetComposicaoDemanda(reservatorioID)
	if err != nil {
		return nil, err
	}

	od, err := uc.repo.GetOfertaDemanda(reservatorioID)
	if err != nil {
		return nil, err
	}

	var listaBM []map[string]interface{}
	for _, item := range bm {
		listaBM = append(listaBM, map[string]interface{}{
			"Mês":               item.Mes,
			"Afluência (m³/s)":  item.AfluenciaM3s,
			"Demanda (m³/s)":    item.DemandasM3s,
			"Balanço (m³/s)":    item.AfluenciaM3s - item.DemandasM3s,
			"Evaporação (m³/s)": item.EvaporacaoM3s,
		})
	}

	var listaCD []map[string]interface{}
	for _, item := range cd {
		listaCD = append(listaCD, map[string]interface{}{
			"Uso":         item.Usos,
			"Vazão (L/s)": item.DemandasHm3 * 1000, // [CORREÇÃO] Conversão aplicada
		})
	}

	var listaOD []map[string]interface{}
	for _, item := range od {
		listaOD = append(listaOD, map[string]interface{}{
			"Cenário":       item.Cenarios,
			"Oferta (L/s)":  item.OfertaLs,  // m³/s -> L/s
			"Demanda (L/s)": item.DemandaLs, // m³/s -> L/s
		})
	}

	return &model.BalancoHidricoResumo{
		BalancoMensal:     listaBM,
		ComposicaoDemanda: listaCD,
		OfertaDemanda:     listaOD,
	}, nil
}

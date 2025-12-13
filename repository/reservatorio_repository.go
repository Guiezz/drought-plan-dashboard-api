package repository

import (
	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

type ReservatorioRepository struct {
	db *gorm.DB
}

func NewReservatorioRepository(db *gorm.DB) *ReservatorioRepository {
	return &ReservatorioRepository{db: db}
}

// 1. Implementa GetReservatorios
func (r *ReservatorioRepository) GetReservatorios() ([]model.Reservatorio, error) {
	var reservatorios []model.Reservatorio
	result := r.db.Order("nome asc").Find(&reservatorios)
	return reservatorios, result.Error
}

// 2. Implementa GetUltimoMonitoramento (O que estava faltando)
func (r *ReservatorioRepository) GetUltimoMonitoramento(reservatorioID int) (*model.Monitoramento, error) {
	var monitoramento model.Monitoramento
	result := r.db.Where("reservatorio_id = ?", reservatorioID).
		Order("data desc").
		First(&monitoramento)

	if result.Error != nil {
		return nil, result.Error
	}
	return &monitoramento, nil
}

func (r *ReservatorioRepository) GetMetas(reservatorioID int) ([]model.VolumeMeta, error) {
	var metas []model.VolumeMeta
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&metas)
	return metas, result.Error
}

func (r *ReservatorioRepository) GetHistoricoMonitoramento(reservatorioID int, limit int) ([]model.Monitoramento, error) {
	var historico []model.Monitoramento
	query := r.db.Where("reservatorio_id = ?", reservatorioID).Order("data asc") // Gráfico geralmente pede ASC

	if limit > 0 {
		query = r.db.Where("reservatorio_id = ?", reservatorioID).Order("data desc").Limit(limit)
	}

	result := query.Find(&historico)
	return historico, result.Error
}

func (r *ReservatorioRepository) GetUsosAgua(reservatorioID int) ([]model.UsoAgua, error) {
	var usos []model.UsoAgua
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&usos)
	return usos, result.Error
}

func (r *ReservatorioRepository) GetResponsaveis(reservatorioID int) ([]model.Responsavel, error) {
	var responsaveis []model.Responsavel
	result := r.db.Where("reservatorio_id = ?", reservatorioID).
		Order("grupo, organizacao, nome").
		Find(&responsaveis)
	return responsaveis, result.Error
}

func (r *ReservatorioRepository) GetReservatorioByID(id int) (*model.Reservatorio, error) {
	var res model.Reservatorio
	result := r.db.First(&res, id)
	return &res, result.Error
}

// Busca todas as datas de monitoramento de um reservatório (para checagem de duplicidade)
func (r *ReservatorioRepository) GetDatasMonitoramento(reservatorioID int) (map[string]bool, error) {
	var datas []string
	// Seleciona apenas a coluna 'data'
	// O Postgres retorna datas como string no formato padrão YYYY-MM-DD
	result := r.db.Model(&model.Monitoramento{}).
		Where("reservatorio_id = ?", reservatorioID).
		Pluck("data", &datas) // Pluck joga os valores direto no slice

	if result.Error != nil {
		return nil, result.Error
	}

	// Converte para Map para busca rápida (O(1))
	mapaDatas := make(map[string]bool)
	for _, d := range datas {
		// As vezes o GORM/Driver traz a data com timestamp (ex: 2023-01-01T00:00:00Z)
		// Vamos pegar apenas os 10 primeiros caracteres (YYYY-MM-DD)
		if len(d) >= 10 {
			mapaDatas[d[:10]] = true
		}
	}

	return mapaDatas, nil
}

// Insere múltiplos registros de uma vez
func (r *ReservatorioRepository) SalvarMonitoramentos(registros []model.Monitoramento) error {
	if len(registros) == 0 {
		return nil
	}
	// Create do GORM aceita um slice e faz um INSERT múltiplo automaticamente
	return r.db.Create(&registros).Error
}

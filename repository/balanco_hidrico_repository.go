package repository

import (
	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

type BalancoHidricoRepository struct {
	db *gorm.DB
}

func NewBalancoHidricoRepository(db *gorm.DB) *BalancoHidricoRepository {
	return &BalancoHidricoRepository{db: db}
}

func (r *BalancoHidricoRepository) GetBalancoMensal(reservatorioID int) ([]model.BalancoMensal, error) {
	var dados []model.BalancoMensal
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&dados)
	return dados, result.Error
}

func (r *BalancoHidricoRepository) GetComposicaoDemanda(reservatorioID int) ([]model.ComposicaoDemanda, error) {
	var dados []model.ComposicaoDemanda
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&dados)
	return dados, result.Error
}

func (r *BalancoHidricoRepository) GetOfertaDemanda(reservatorioID int) ([]model.OfertaDemanda, error) {
	var dados []model.OfertaDemanda
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&dados)
	return dados, result.Error
}

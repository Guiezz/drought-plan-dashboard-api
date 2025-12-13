package repository

import (
	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

type ResponsavelRepository struct {
	db *gorm.DB
}

func NewResponsavelRepository(db *gorm.DB) *ResponsavelRepository {
	return &ResponsavelRepository{db: db}
}

func (r *ResponsavelRepository) Listar(reservatorioID int) ([]model.Responsavel, error) {
	var responsaveis []model.Responsavel
	result := r.db.Where("reservatorio_id = ?", reservatorioID).
		Order("grupo, organizacao, nome").
		Find(&responsaveis)
	return responsaveis, result.Error
}

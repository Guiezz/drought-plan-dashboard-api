package repository

import (
	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

type UsoAguaRepository struct {
	db *gorm.DB
}

func NewUsoAguaRepository(db *gorm.DB) *UsoAguaRepository {
	return &UsoAguaRepository{db: db}
}

func (r *UsoAguaRepository) Listar(reservatorioID int) ([]model.UsoAgua, error) {
	var usos []model.UsoAgua
	result := r.db.Where("reservatorio_id = ?", reservatorioID).Find(&usos)
	return usos, result.Error
}

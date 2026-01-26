package repository

import (
	"github.com/guiezz/dashboard-api/model/simulador"
	"gorm.io/gorm"
)

type SimulacaoRepository struct {
	db *gorm.DB
}

func NewSimulacaoRepository(db *gorm.DB) *SimulacaoRepository {
	return &SimulacaoRepository{db: db}
}

// Busca os dados cadastrais do Açude de Simulação pelo COD (ID)
func (r *SimulacaoRepository) GetAcude(cod int) (*simulador.SimAcude, error) {
	var acude simulador.SimAcude
	result := r.db.First(&acude, "codigo = ?", cod)
	return &acude, result.Error
}

// Busca a curva Cota-Área-Volume ordenada
func (r *SimulacaoRepository) GetCAV(acudeID int) ([]simulador.SimCAV, error) {
	var cav []simulador.SimCAV
	result := r.db.Where("acude_id = ?", acudeID).Order("volume asc").Find(&cav)
	return cav, result.Error
}

// Busca as vazões históricas dentro de um intervalo de anos
func (r *SimulacaoRepository) GetVazoes(acudeID, anoInicio, anoFim int) ([]simulador.SimVazao, error) {
	var vazoes []simulador.SimVazao
	result := r.db.Where("acude_id = ? AND ano >= ? AND ano <= ?", acudeID, anoInicio, anoFim).
		Order("ano asc").
		Find(&vazoes)
	return vazoes, result.Error
}

// Busca os dados de evaporação pelo ID da estação (que vem do cadastro do açude)
func (r *SimulacaoRepository) GetEvaporacao(estacaoID int) (*simulador.SimEvaporacao, error) {
	var evap simulador.SimEvaporacao
	result := r.db.First(&evap, "codigo = ?", estacaoID)
	return &evap, result.Error
}

// Lista todos os açudes disponíveis para simulação (para o dropdown do frontend)
func (r *SimulacaoRepository) ListarAcudes() ([]simulador.SimAcude, error) {
	var acudes []simulador.SimAcude
	result := r.db.Select("codigo", "nome", "municipio", "capacidade").Order("nome asc").Find(&acudes)
	return acudes, result.Error
}

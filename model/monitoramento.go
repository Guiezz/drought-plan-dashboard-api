package model

import "time"

type Monitoramento struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Data             time.Time `gorm:"index" json:"data"`
	VolumeHm3        float64   `gorm:"column:volume_hm3" json:"volume_hm3"`
	VolumePercentual float64   `gorm:"column:volume_percentual" json:"volume_percentual"`

	ReservatorioID uint `json:"reservatorio_id"`
}

type HistoricoTabela struct {
	Data       string  `json:"Data"`           // Ex: "25/12/2023"
	EstadoSeca string  `json:"Estado de Seca"` // Com espaços
	VolumeHm3  float64 `json:"Volume (hm3)"`   // Com unidade e parênteses
}

func (Monitoramento) TableName() string {
	return "monitoramento"
}

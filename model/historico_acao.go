package model

import "time"

type HistoricoAcao struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	PlanoAcaoID      uint      `gorm:"not null;index" json:"plano_acao_id"`
	UsuarioID        uint      `gorm:"not null" json:"usuario_id"`
	SituacaoAnterior string    `json:"situacao_anterior"`
	SituacaoNova     string    `json:"situacao_nova"`
	DataAlteracao    time.Time `json:"data_alteracao"`

	Usuario Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
}

func (HistoricoAcao) TableName() string {
	return "historico_acoes"
}

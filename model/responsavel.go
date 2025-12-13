package model

type Responsavel struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ReservatorioID uint   `json:"reservatorio_id"`
	Nome           string `json:"nome"`
	Grupo          string `json:"grupo"`
	Organizacao    string `json:"organizacao"`
	Cargo          string `json:"cargo"`
}

func (Responsavel) TableName() string {
	return "responsaveis" // Note o plural, conforme o script Python criou
}

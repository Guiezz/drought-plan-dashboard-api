package model

type VolumeMeta struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	MesNum  int    `json:"mes_num"`
	MesNome string `json:"mes_nome"`

	Meta1v float64 `gorm:"column:meta1v" json:"meta1v"`
	Meta2v float64 `gorm:"column:meta2v" json:"meta2v"`
	Meta3v float64 `gorm:"column:meta3v" json:"meta3v"`

	ReservatorioID uint `json:"reservatorio_id"`
}

func (VolumeMeta) TableName() string {
	return "volume_meta"
}

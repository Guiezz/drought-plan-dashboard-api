package model

type UsoAgua struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	ReservatorioID uint    `json:"reservatorio_id"`
	Uso            string  `json:"uso"`
	VazaoNormal    float64 `json:"vazao_normal"`
	VazaoEscassez  float64 `json:"vazao_escassez"`
}

func (UsoAgua) TableName() string {
	return "uso_agua"
}

package model

type BalancoMensal struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	ReservatorioID uint `json:"reservatorio_id"`

	Mes           string  `gorm:"column:mes" json:"mes"`
	AfluenciaM3s  float64 `gorm:"column:afluencia_m3s" json:"afluencia_m3s"`
	DemandasM3s   float64 `gorm:"column:demandas_m3s" json:"demandas_m3s"`
	EvaporacaoM3s float64 `gorm:"column:evaporacao_m3s" json:"evaporacao_m3s"`
}

func (BalancoMensal) TableName() string {
	return "balanco_mensal"
}

type ComposicaoDemanda struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	ReservatorioID uint `json:"reservatorio_id"`

	Usos        string  `gorm:"column:usos" json:"usos"`
	DemandasHm3 float64 `gorm:"column:demandas_hm3" json:"demandas_hm3"`
}

func (ComposicaoDemanda) TableName() string {
	return "composicao_demanda"
}

type OfertaDemanda struct {
	ID             uint `gorm:"primaryKey" json:"id"`
	ReservatorioID uint `json:"reservatorio_id"`

	Cenarios   string  `gorm:"column:cenarios" json:"cenarios"`
	OfertaM3s  float64 `gorm:"column:oferta_m3s" json:"oferta_m3s"`
	DemandaM3s float64 `gorm:"column:demanda_m3s" json:"demanda_m3s"`
}

func (OfertaDemanda) TableName() string {
	return "oferta_demanda"
}

package model

type GatilhoMensalPGPS struct {
	MesNum       int     `json:"mes_num"`
	MesNome      string  `json:"mes_nome"`
	SecaSeveraHm3 float64 `json:"seca_severa_hm3"`
	SecaHm3       float64 `json:"seca_hm3"`
	AlertaHm3     float64 `json:"alerta_hm3"`
	NormalHm3     float64 `json:"normal_hm3"`
}

type GatilhosPGPSResponse struct {
	ReservatorioID  uint                `json:"reservatorio_id"`
	NomeReservatorio string             `json:"nome_reservatorio"`
	CapacidadeHm3   float64             `json:"capacidade_hm3"`
	Gatilhos        []GatilhoMensalPGPS `json:"gatilhos"`
}

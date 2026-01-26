package simulador

import "time"

// CotaAreaVolume representa a tabela CAV que antes estava no Excel
type CotaAreaVolume struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	ReservatorioID uint    `json:"reservatorio_id"`
	Cota           float64 `json:"cota"`
	Area           float64 `json:"area"`   // Em km² (conforme seu código Python)
	Volume         float64 `json:"volume"` // Em hm³
}

func (CotaAreaVolume) TableName() string {
	return "cota_area_volume"
}

// SimulacaoRequest define o que o front/usuário envia para simular
type SimulacaoRequest struct {
	ReservatorioID uint      `json:"reservatorio_id"`
	VolumeInicial  float64   `json:"volume_inicial"` // Em hm³
	DataInicio     time.Time `json:"data_inicio"`
	DataFim        time.Time `json:"data_fim"`

	// Se true, usa médias históricas do banco. Se false, usa os arrays abaixo.
	UsarMediaHistorica bool `json:"usar_media_historica"`

	// Opcionais (se quiser simular cenários manuais)
	DemandasMensais []float64 `json:"demandas_mensais"` // 12 valores (m³/s)
}

// SimulacaoResultados é o JSON de resposta
type SimulacaoResultado struct {
	Data       string  `json:"data"`
	Volume     float64 `json:"volume_hm3"`
	Afluencia  float64 `json:"afluencia_hm3"`
	Retirada   float64 `json:"retirada_hm3"`
	Evaporacao float64 `json:"evaporacao_hm3"`
	Vertimento float64 `json:"vertimento_hm3"`
	Alerta     string  `json:"alerta,omitempty"`
}

type SimulacaoResponse struct {
	Resultados            []SimulacaoResultado `json:"resultados"`
	FrequenciaNaoAtendida float64              `json:"frequencia_nao_atendida"`
	VolumeFinal           float64              `json:"volume_final"`
}

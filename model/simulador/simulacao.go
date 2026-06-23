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

	// Multi-cenário: se preenchido, executa simulação para cada cenário
	Cenarios []CenarioSimulacao `json:"cenarios,omitempty"`
}

// CenarioSimulacao define um cenário histórico para simulação multi-cenário
type CenarioSimulacao struct {
	Nome string `json:"nome"` // "Seca 1915", "2012-2017", etc.
	Anos []int  `json:"anos"` // [1915] ou [2012,2013,2014,2015,2016,2017]
}

// SimulacaoResultados é o JSON de resposta
type SimulacaoResultado struct {
	Data          string  `json:"data"`
	VolumeInicial float64 `json:"volume_inicial_hm3"`
	VolumeFinal   float64 `json:"volume_final_hm3"`
	Afluencia     float64 `json:"afluencia_hm3"`
	Retirada      float64 `json:"retirada_hm3"`
	Evaporacao    float64 `json:"evaporacao_hm3"`
	Vertimento    float64 `json:"vertimento_hm3"`
	Alerta        string  `json:"alerta,omitempty"`
}

type SimulacaoResponse struct {
	Resultados            []SimulacaoResultado `json:"resultados"`
	FrequenciaNaoAtendida float64              `json:"frequencia_nao_atendida"`
	VolumeFinal           float64              `json:"volume_final"`
}

// ResultadoCenario é o resultado de UM cenário na simulação multi-cenário
type ResultadoCenario struct {
	Nome                  string                `json:"nome"`
	Resultados            []SimulacaoResultado  `json:"resultados"`
	FrequenciaNaoAtendida float64               `json:"frequencia_nao_atendida"`
	VolumeFinal           float64               `json:"volume_final"`
}

// EstatisticaDescritiva agrega métricas de um conjunto de valores
type EstatisticaDescritiva struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Media   float64 `json:"media"`
	Mediana float64 `json:"mediana"`
	P10     float64 `json:"p10"`
	P90     float64 `json:"p90"`
}

// DistribuicaoResultados consolida a distribuição dos resultados entre cenários
type DistribuicaoResultados struct {
	FrequenciaNaoAtendida EstatisticaDescritiva `json:"frequencia_nao_atendida"`
	VolumeFinal           EstatisticaDescritiva `json:"volume_final"`
}

// SimulacaoMultiResponse é a resposta da simulação multi-cenário
type SimulacaoMultiResponse struct {
	Cenarios     []ResultadoCenario    `json:"cenarios"`
	Distribuicao DistribuicaoResultados `json:"distribuicao"`
}

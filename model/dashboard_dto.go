package model

type DashboardResumo struct {
	VolumeAtualHm3      float64           `json:"volumeAtualHm3"`
	VolumePercentual    float64           `json:"volumePercentual"`
	EstadoAtualSeca     string            `json:"estadoAtualSeca"`
	DataUltimaMedicao   string            `json:"dataUltimaMedicao"`
	DiasDesdeMudanca    int               `json:"diasDesdeUltimaMudanca"`
	MedidasRecomendadas []PlanoAcaoResumo `json:"medidasRecomendadas"`
}

type PlanoAcaoResumo struct {
	Acao         string `json:"acao"`
	Descricao    string `json:"descricao"`
	Responsaveis string `json:"responsaveis"`
}

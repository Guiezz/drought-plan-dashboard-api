package model

type GraficoVolumeData struct {
	Data   string  `json:"Data"` // Frontend espera "Data" com D maiúsculo
	Volume float64 `json:"volume"`
	Meta1  float64 `json:"meta1"`
	Meta2  float64 `json:"meta2"`
	Meta3  float64 `json:"meta3"`
}

type BalancoHidricoResumo struct {
	BalancoMensal     []map[string]interface{} `json:"balancoMensal"`
	ComposicaoDemanda []map[string]interface{} `json:"composicaoDemanda"`
	OfertaDemanda     []map[string]interface{} `json:"ofertaDemanda"`
}

package model

// Estrutura que mapeia a resposta JSON da API da FUNCEME
type FuncemeResponse struct {
	Data struct {
		List []FuncemeRegistro `json:"list"`
	} `json:"data"`
}

type FuncemeRegistro struct {
	DataStr    string  `json:"data"` // Vem como string "2023-10-25"
	Volume     float64 `json:"volume"`
	VolumePerc float64 `json:"volume_perc"`
}

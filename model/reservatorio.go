package model

type Reservatorio struct {
	ID             uint    `gorm:"primaryKey" json:"id"`
	Nome           string  `gorm:"uniqueIndex;not null" json:"nome"`
	Municipio      string  `json:"municipio"`
	Descricao      string  `json:"descricao"`
	Lat            float64 `json:"lat"`
	Long           float64 `json:"long"`
	NomeImagem     string  `json:"nome_imagem"`
	NomeImagemUsos string  `json:"nome_imagem_usos"`
	CodigoFunceme  string  `json:"codigo_funceme"`

	// Relacionamentos (Opcional, mas útil no GORM)
	Monitoramentos []Monitoramento `gorm:"foreignKey:ReservatorioID" json:"-"`
	Metas          []VolumeMeta    `gorm:"foreignKey:ReservatorioID" json:"-"`
}

type ReservatorioDetalhes struct {
	ID            uint    `json:"id"`
	Nome          string  `json:"nome"`
	Municipio     string  `json:"municipio"`
	Descricao     string  `json:"descricao"`
	Lat           float64 `json:"lat"`
	Long          float64 `json:"long"`
	UrlImagem     string  `json:"url_imagem"`
	UrlImagemUsos string  `json:"url_imagem_usos"`
}

func (Reservatorio) TableName() string {
	return "reservatorio"
}

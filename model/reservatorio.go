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
	Capacidadehm3  float64 `json:"capacidade_hm3"`
	CodigoFunceme  string  `json:"codigo_funceme"`

	// Ficha técnica (opcional; carregada de dados_importacao/ficha_tecnica.csv)
	Bacia               *string  `json:"bacia,omitempty"`
	AreaDrenagemKm2     *float64 `json:"area_drenagem_km2,omitempty"`
	AnoConstrucao       *int     `json:"ano_construcao,omitempty"`
	Construtor          *string  `json:"construtor,omitempty"`
	RioBarrado          *string  `json:"rio_barrado,omitempty"`
	MunicipiosAtendidos *string  `json:"municipios_atendidos,omitempty"`

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
	CapacidadeHm3 float64 `json:"capacidade_hm3,omitempty"`
	CodigoFunceme string  `json:"codigo_funceme,omitempty"`

	Bacia               *string  `json:"bacia,omitempty"`
	AreaDrenagemKm2     *float64 `json:"area_drenagem_km2,omitempty"`
	AnoConstrucao       *int     `json:"ano_construcao,omitempty"`
	Construtor          *string  `json:"construtor,omitempty"`
	RioBarrado          *string  `json:"rio_barrado,omitempty"`
	MunicipiosAtendidos *string  `json:"municipios_atendidos,omitempty"`
}

func (Reservatorio) TableName() string {
	return "reservatorio"
}

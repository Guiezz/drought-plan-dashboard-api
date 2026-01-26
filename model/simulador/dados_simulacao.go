package simulador

// Tabela 'acudes' do Excel
type SimAcude struct {
	Codigo        int     `gorm:"primaryKey;autoIncrement:false" json:"codigo"` // Coluna COD
	Nome          string  `json:"nome"`                                         // Coluna CORPO
	Capacidade    float64 `json:"capacidade_m3"`                                // Coluna CAPAC (m³)
	Municipio     string  `json:"municipio"`                                    // Coluna MUNICIPIO
	EstacaoEvapID int     `json:"estacao_evap_id"`                              // Coluna Est. Evap.

	// Relacionamentos
	Vazoes []SimVazao `gorm:"foreignKey:AcudeID;references:Codigo" json:"-"`
	CAV    []SimCAV   `gorm:"foreignKey:AcudeID;references:Codigo" json:"-"`
}

func (SimAcude) TableName() string { return "sim_acudes" }

// Tabela 'cav' do Excel
type SimCAV struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	AcudeID int     `gorm:"index" json:"acude_id"` // Coluna COD
	Cota    float64 `json:"cota"`                  // Coluna COTA
	Area    float64 `json:"area"`                  // Coluna area
	Volume  float64 `json:"volume"`                // Coluna volume
}

func (SimCAV) TableName() string { return "sim_cav" }

// Tabela 'evaporacao' do Excel
type SimEvaporacao struct {
	Codigo    int     `gorm:"primaryKey;autoIncrement:false" json:"codigo"` // Coluna COD
	Municipio string  `json:"municipio"`
	Jan       float64 `json:"jan"`
	Fev       float64 `json:"fev"`
	Mar       float64 `json:"mar"`
	Abr       float64 `json:"abr"`
	Mai       float64 `json:"mai"`
	Jun       float64 `json:"jun"`
	Jul       float64 `json:"jul"`
	Ago       float64 `json:"ago"`
	Set       float64 `json:"set"`
	Out       float64 `json:"out"`
	Nov       float64 `json:"nov"`
	Dez       float64 `json:"dez"`
}

func (SimEvaporacao) TableName() string { return "sim_evaporacao" }

// Tabela 'vazoes' do Excel
type SimVazao struct {
	ID      uint    `gorm:"primaryKey" json:"id"`
	AcudeID int     `gorm:"index" json:"acude_id"` // Coluna COD
	Ano     int     `json:"ano"`                   // Coluna ANO
	Jan     float64 `json:"jan"`
	Fev     float64 `json:"fev"`
	Mar     float64 `json:"mar"`
	Abr     float64 `json:"abr"`
	Mai     float64 `json:"mai"`
	Jun     float64 `json:"jun"`
	Jul     float64 `json:"jul"`
	Ago     float64 `json:"ago"`
	Set     float64 `json:"set"`
	Out     float64 `json:"out"`
	Nov     float64 `json:"nov"`
	Dez     float64 `json:"dez"`
}

func (SimVazao) TableName() string { return "sim_vazoes" }

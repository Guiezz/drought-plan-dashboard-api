package model

type PlanoAcao struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	ReservatorioID   uint   `json:"reservatorio_id"`
	EstadoSeca       string `json:"estado_seca"`
	Problemas        string `json:"problemas"`
	TiposImpactos    string `json:"tipos_impactos"`
	Acoes            string `json:"acoes"`
	DescricaoAcao    string `json:"descricao_acao"`
	ClassesAcao      string `json:"classes_acao"`
	Responsaveis     string `json:"responsaveis"`
	Situacao         string `json:"situacao"`
	Indicadores      string `json:"indicadores"`
	OrgaosEnvolvidos string `json:"orgaos_envolvidos"`
}

// TableName garante o mapeamento correto com a tabela criada pelo Python
func (PlanoAcao) TableName() string {
	return "plano_acao"
}

// Struct auxiliar para os filtros do frontend
type FiltrosPlanoAcao struct {
	Estados   []string `json:"estados"`
	Impactos  []string `json:"impactos"`
	Problemas []string `json:"problemas"`
	Acoes     []string `json:"acoes"`
}

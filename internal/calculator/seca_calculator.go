package calculator

import (
	"github.com/guiezz/dashboard-api/model"
)

// Define a interface para facilitar testes
type SecaCalculator interface {
	CalcularEstado(m *model.Monitoramento, metas []model.VolumeMeta) string
	CalcularDiasDesdeMudanca(estadoAtual string, historico []model.Monitoramento, metas []model.VolumeMeta) int
}

type secaCalculator struct{}

// NewSecaCalculator cria uma nova instância da calculadora
func NewSecaCalculator() SecaCalculator {
	return &secaCalculator{}
}

// CalcularEstado determina a situação do reservatório com base no volume e nas metas do mês
func (c *secaCalculator) CalcularEstado(m *model.Monitoramento, metas []model.VolumeMeta) string {
	if m == nil {
		return "NÃO DEFINIDO"
	}

	mesMonitoramento := int(m.Data.Month())

	// Encontra a meta correspondente ao mês do monitoramento
	var metaMes model.VolumeMeta
	found := false
	for _, meta := range metas {
		if meta.MesNum == mesMonitoramento {
			metaMes = meta
			found = true
			break
		}
	}

	if !found {
		return "NÃO DEFINIDO"
	}

	// Lógica de faixas de volume (assume percentual de 0 a 100)
	volDecimal := m.VolumePercentual / 100.0

	if volDecimal < metaMes.Meta1v {
		return "SECA SEVERA"
	}
	if volDecimal < metaMes.Meta2v {
		return "SECA"
	}
	if volDecimal < metaMes.Meta3v {
		return "ALERTA"
	}
	return "NORMAL"
}

// CalcularDiasDesdeMudanca percorre o histórico para achar quando o estado mudou
func (c *secaCalculator) CalcularDiasDesdeMudanca(estadoAtual string, historico []model.Monitoramento, metas []model.VolumeMeta) int {
	if len(historico) == 0 {
		return 0
	}

	// Assume que historico[0] é o registro mais recente (ordem DESC)
	dataAtual := historico[0].Data

	for _, registro := range historico {
		// Reutiliza a função de cálculo interna (método deste struct)
		estadoRegistro := c.CalcularEstado(&registro, metas)

		// Se o estado for diferente do atual, encontramos a data da mudança
		if estadoRegistro != estadoAtual {
			// Diferença em dias
			diff := dataAtual.Sub(registro.Data)
			return int(diff.Hours() / 24)
		}
	}

	// Se percorreu todo o histórico e não mudou, retorna dias totais do histórico disponível
	diff := dataAtual.Sub(historico[len(historico)-1].Data)
	return int(diff.Hours() / 24)
}

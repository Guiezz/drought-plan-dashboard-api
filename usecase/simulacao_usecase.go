package usecase

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/model/simulador"
)

type SimulacaoRepositoryInterface interface {
	GetAcude(cod int) (*simulador.SimAcude, error)
	GetCAV(acudeID int) ([]simulador.SimCAV, error)
	GetEvaporacao(estacaoID int) (*simulador.SimEvaporacao, error)
	GetVazoes(acudeID, anoInicio, anoFim int) ([]simulador.SimVazao, error)
	GetVazoesByAnos(acudeID int, anos []int) ([]simulador.SimVazao, error)
	GetAnosVazoes(acudeID int) ([]int, error)
	ListarAcudes() ([]simulador.SimAcude, error)
}

type SimulacaoUseCase struct {
	repo SimulacaoRepositoryInterface
	calc *calculator.SimuladorHidrico
}

func NewSimulacaoUseCase(repo SimulacaoRepositoryInterface, calc *calculator.SimuladorHidrico) *SimulacaoUseCase {
	return &SimulacaoUseCase{
		repo: repo,
		calc: calc,
	}
}

func (u *SimulacaoUseCase) Executar(req simulador.SimulacaoRequest) (*simulador.SimulacaoResponse, error) {
	// 1. Validar datas
	if req.DataFim.Before(req.DataInicio) {
		return nil, errors.New("data final deve ser posterior à data inicial")
	}

	// 2. Buscar dados do Açude
	acude, err := u.repo.GetAcude(int(req.ReservatorioID))
	if err != nil {
		return nil, errors.New("açude não encontrado na base de simulação")
	}

	// 3. Buscar CAV
	cavData, err := u.repo.GetCAV(acude.Codigo)
	if err != nil || len(cavData) < 3 {
		return nil, errors.New("curva Cota-Área-Volume insuficiente")
	}

	// 4. Buscar Evaporação
	evapData, err := u.repo.GetEvaporacao(acude.EstacaoEvapID)
	if err != nil {
		return nil, errors.New("dados de evaporação não encontrados")
	}
	evapMensal := []float64{
		evapData.Jan, evapData.Fev, evapData.Mar, evapData.Abr, evapData.Mai, evapData.Jun,
		evapData.Jul, evapData.Ago, evapData.Set, evapData.Out, evapData.Nov, evapData.Dez,
	}

	// 5. Montar Séries Temporais (Afluencia, Demanda e Evaporação) com DIAS EXATOS
	var afluencias []float64
	var demandas []float64
	var evaporacaoExpandida []float64

	// Pega vazões do banco
	anoInicio := req.DataInicio.Year()
	anoFim := req.DataFim.Year()
	vazoes, err := u.repo.GetVazoes(acude.Codigo, anoInicio, anoFim)
	if err != nil {
		return nil, err
	}

	// Mapeia vazões
	mapaVazoes := make(map[int][]float64)
	for _, v := range vazoes {
		mapaVazoes[v.Ano] = []float64{
			v.Jan, v.Fev, v.Mar, v.Abr, v.Mai, v.Jun,
			v.Jul, v.Ago, v.Set, v.Out, v.Nov, v.Dez,
		}
	}

	currDate := req.DataInicio

	// Loop Principal: Processa mês a mês considerando dias reais
	for !currDate.After(req.DataFim) {
		mesIdx := int(currDate.Month()) - 1
		ano := currDate.Year()

		// --- CÁLCULO DOS DIAS EXATOS DO MÊS ---
		proximoMes := currDate.AddDate(0, 1, 0)
		diasNoMes := proximoMes.Sub(currDate).Hours() / 24
		segundosMes := diasNoMes * 24 * 60 * 60 // Segundos exatos neste mês (ex: Fev=28/29, Jul=31)

		// 1. Afluência (m³/s -> hm³)
		vazaoMes := 0.0
		if vals, ok := mapaVazoes[ano]; ok {
			vazaoMes = vals[mesIdx]
		}
		afluenciaHm3 := (vazaoMes * segundosMes) / 1e6
		afluencias = append(afluencias, afluenciaHm3)

		// 2. Demanda (m³/s -> hm³)
		demandaM3s := 0.0
		if len(req.DemandasMensais) == 12 {
			demandaM3s = req.DemandasMensais[mesIdx]
		} else if len(req.DemandasMensais) == 1 {
			demandaM3s = req.DemandasMensais[0]
		}
		demandaHm3 := (demandaM3s * segundosMes) / 1e6
		demandas = append(demandas, demandaHm3)

		// 3. Evaporação
		evaporacaoExpandida = append(evaporacaoExpandida, evapMensal[mesIdx])

		// Avança data
		currDate = proximoMes
	}

	// 7. Chamar Calculadora
	// Converte capacidade de m³ para hm³
	capacidadeHm3 := acude.Capacidade / 1e6

	resultado := u.calc.Simular(
		req.VolumeInicial,
		cavData,
		afluencias,
		demandas,
		evaporacaoExpandida,
		capacidadeHm3,
	)

	// 8. Ajustar datas no resultado final
	currDate = req.DataInicio
	for i := range resultado.Resultados {
		resultado.Resultados[i].Data = currDate.Format("2006-01")
		currDate = currDate.AddDate(0, 1, 0)
	}

	return &resultado, nil
}

func (u *SimulacaoUseCase) ListarOpcoes() ([]simulador.SimAcude, error) {
	return u.repo.ListarAcudes()
}

func (u *SimulacaoUseCase) ListarAnos(reservatorioID int) ([]int, error) {
	return u.repo.GetAnosVazoes(reservatorioID)
}

// ExecutarMultiCenario executa a simulação para múltiplos cenários históricos
func (u *SimulacaoUseCase) ExecutarMultiCenario(req simulador.SimulacaoRequest) (*simulador.SimulacaoMultiResponse, error) {
	// 1. Validar datas
	if req.DataFim.Before(req.DataInicio) {
		return nil, errors.New("data final deve ser posterior à data inicial")
	}

	// 2. Buscar dados do Açude
	acude, err := u.repo.GetAcude(int(req.ReservatorioID))
	if err != nil {
		return nil, errors.New("açude não encontrado na base de simulação")
	}

	// 3. Buscar CAV
	cavData, err := u.repo.GetCAV(acude.Codigo)
	if err != nil || len(cavData) < 3 {
		return nil, errors.New("curva Cota-Área-Volume insuficiente")
	}

	// 4. Buscar Evaporação
	evapData, err := u.repo.GetEvaporacao(acude.EstacaoEvapID)
	if err != nil {
		return nil, errors.New("dados de evaporação não encontrados")
	}
	evapMensal := []float64{
		evapData.Jan, evapData.Fev, evapData.Mar, evapData.Abr, evapData.Mai, evapData.Jun,
		evapData.Jul, evapData.Ago, evapData.Set, evapData.Out, evapData.Nov, evapData.Dez,
	}

	// 5. Coletar TODOS os anos únicos de todos os cenários (uma única query)
	anosUnicos := make(map[int]struct{})
	for _, c := range req.Cenarios {
		for _, ano := range c.Anos {
			anosUnicos[ano] = struct{}{}
		}
	}
	anosLista := make([]int, 0, len(anosUnicos))
	for ano := range anosUnicos {
		anosLista = append(anosLista, ano)
	}

	// 6. Buscar vazões de uma vez
	vazoes, err := u.repo.GetVazoesByAnos(acude.Codigo, anosLista)
	if err != nil {
		return nil, err
	}

	// 7. Mapear vazões por ano
	mapaVazoes := make(map[int][]float64)
	for _, v := range vazoes {
		mapaVazoes[v.Ano] = []float64{
			v.Jan, v.Fev, v.Mar, v.Abr, v.Mai, v.Jun,
			v.Jul, v.Ago, v.Set, v.Out, v.Nov, v.Dez,
		}
	}

	// 8. Séries compartilhadas (demanda e evaporação)
	demandas, evaporacaoExpandida := buildSeriesDemandaEvap(req, evapMensal)

	capacidadeHm3 := acude.Capacidade / 1e6

	// 9. Executar simulação para cada cenário
	resultadosCenarios := make([]simulador.ResultadoCenario, len(req.Cenarios))
	for i, cenario := range req.Cenarios {
		afluencias := buildSeriesAfluencia(cenario, mapaVazoes, req.DataInicio, req.DataFim)

		resultado := u.calc.Simular(
			req.VolumeInicial,
			cavData,
			afluencias,
			demandas,
			evaporacaoExpandida,
			capacidadeHm3,
		)

		resultadosCenarios[i] = simulador.ResultadoCenario{
			Nome:                  cenario.Nome,
			Resultados:            resultado.Resultados,
			FrequenciaNaoAtendida: resultado.FrequenciaNaoAtendida,
			VolumeFinal:           resultado.VolumeFinal,
		}
	}

	// 10. Ajustar datas em todos os resultados
	for i := range resultadosCenarios {
		currDate := req.DataInicio
		for j := range resultadosCenarios[i].Resultados {
			resultadosCenarios[i].Resultados[j].Data = currDate.Format("2006-01")
			currDate = currDate.AddDate(0, 1, 0)
		}
	}

	// 11. Calcular distribuição
	freqValues := make([]float64, len(resultadosCenarios))
	volFinalValues := make([]float64, len(resultadosCenarios))
	for i, r := range resultadosCenarios {
		freqValues[i] = r.FrequenciaNaoAtendida
		volFinalValues[i] = r.VolumeFinal
	}

	return &simulador.SimulacaoMultiResponse{
		Cenarios: resultadosCenarios,
		Distribuicao: simulador.DistribuicaoResultados{
			FrequenciaNaoAtendida: calcularEstatisticas(freqValues),
			VolumeFinal:           calcularEstatisticas(volFinalValues),
		},
	}, nil
}

// buildSeriesAfluencia monta a série de afluências para um cenário, ciclando os anos históricos
func buildSeriesAfluencia(cenario simulador.CenarioSimulacao, mapaVazoes map[int][]float64, dataInicio, dataFim time.Time) []float64 {
	var afluencias []float64
	currDate := dataInicio
	nAnos := len(cenario.Anos)

	for !currDate.After(dataFim) {
		mesIdx := int(currDate.Month()) - 1
		totalMeses := len(afluencias)

		// Cicla pelos anos do cenário: yearIndex = (totalMeses / 12) % nAnos
		yearIndex := (totalMeses / 12) % nAnos
		anoHistorico := cenario.Anos[yearIndex]

		vazaoMes := 0.0
		if vals, ok := mapaVazoes[anoHistorico]; ok {
			vazaoMes = vals[mesIdx]
		}

		proximoMes := currDate.AddDate(0, 1, 0)
		diasNoMes := proximoMes.Sub(currDate).Hours() / 24
		segundosMes := diasNoMes * 24 * 60 * 60

		afluenciaHm3 := (vazaoMes * segundosMes) / 1e6
		afluencias = append(afluencias, afluenciaHm3)

		currDate = proximoMes
	}

	return afluencias
}

// buildSeriesDemandaEvap monta as séries de demanda e evaporação (compartilhadas entre cenários)
func buildSeriesDemandaEvap(req simulador.SimulacaoRequest, evapMensal []float64) (demandas, evaporacaoExpandida []float64) {
	currDate := req.DataInicio
	for !currDate.After(req.DataFim) {
		mesIdx := int(currDate.Month()) - 1

		proximoMes := currDate.AddDate(0, 1, 0)
		diasNoMes := proximoMes.Sub(currDate).Hours() / 24
		segundosMes := diasNoMes * 24 * 60 * 60

		demandaM3s := 0.0
		if len(req.DemandasMensais) == 12 {
			demandaM3s = req.DemandasMensais[mesIdx]
		} else if len(req.DemandasMensais) == 1 {
			demandaM3s = req.DemandasMensais[0]
		}
		demandaHm3 := (demandaM3s * segundosMes) / 1e6
		demandas = append(demandas, demandaHm3)

		evaporacaoExpandida = append(evaporacaoExpandida, evapMensal[mesIdx])

		currDate = proximoMes
	}
	return
}

// calcularEstatisticas calcula min, max, média, mediana, P10 e P90 de um conjunto de valores
func calcularEstatisticas(valores []float64) simulador.EstatisticaDescritiva {
	if len(valores) == 0 {
		return simulador.EstatisticaDescritiva{}
	}

	sorted := make([]float64, len(valores))
	copy(sorted, valores)
	sort.Float64s(sorted)

	n := len(sorted)
	soma := 0.0
	for _, v := range sorted {
		soma += v
	}

	return simulador.EstatisticaDescritiva{
		Min:     sorted[0],
		Max:     sorted[n-1],
		Media:   soma / float64(n),
		Mediana: percentil(sorted, 0.5),
		P10:     percentil(sorted, 0.1),
		P90:     percentil(sorted, 0.9),
	}
}

// percentil calcula o percentil p (0.0 a 1.0) via interpolação linear
func percentil(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	idx := p * float64(len(sorted)-1)
	lower := int(math.Floor(idx))
	upper := int(math.Ceil(idx))
	if lower == upper {
		return sorted[lower]
	}
	frac := idx - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

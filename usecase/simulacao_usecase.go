package usecase

import (
	"errors"
	// Importante adicionar math para o cálculo da capacidade
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/guiezz/dashboard-api/repository"
)

type SimulacaoUseCase struct {
	repo *repository.SimulacaoRepository
	calc *calculator.SimuladorHidrico
}

func NewSimulacaoUseCase(repo *repository.SimulacaoRepository, calc *calculator.SimuladorHidrico) *SimulacaoUseCase {
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

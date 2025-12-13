package usecase

import (
	"fmt"
	"os"

	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/model"
)

// Interfaces definem o contrato. O UseCase não sabe que é GORM, só sabe que alguém vai cumprir isso.
type ReservatorioRepositoryInterface interface {
	GetReservatorios() ([]model.Reservatorio, error)
	GetReservatorioByID(id int) (*model.Reservatorio, error)
	GetUltimoMonitoramento(reservatorioID int) (*model.Monitoramento, error)
	GetMetas(reservatorioID int) ([]model.VolumeMeta, error)
	GetHistoricoMonitoramento(reservatorioID int, limit int) ([]model.Monitoramento, error)

	// --- NOVOS MÉTODOS ---
	GetPlanosAcao(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error)
	GetFiltrosPlanoAcao(reservatorioID int) (*model.FiltrosPlanoAcao, error)
	GetUsosAgua(reservatorioID int) ([]model.UsoAgua, error)
	GetResponsaveis(reservatorioID int) ([]model.Responsavel, error)

	GetBalancoMensal(reservatorioID int) ([]model.BalancoMensal, error)
	GetComposicaoDemanda(reservatorioID int) ([]model.ComposicaoDemanda, error)
	GetOfertaDemanda(reservatorioID int) ([]model.OfertaDemanda, error)

	GetDatasMonitoramento(reservatorioID int) (map[string]bool, error)
	SalvarMonitoramentos(registros []model.Monitoramento) error
}

type ReservatorioUseCase struct {
	repo       ReservatorioRepositoryInterface
	calc       calculator.SecaCalculator
	funcemeSvc funceme.Service
}

func NewReservatorioUseCase(repo ReservatorioRepositoryInterface, calc calculator.SecaCalculator, funcemeSvc funceme.Service) *ReservatorioUseCase {
	return &ReservatorioUseCase{
		repo:       repo,
		calc:       calc,
		funcemeSvc: funcemeSvc,
	}
}

func (uc *ReservatorioUseCase) ListarTodos() ([]model.Reservatorio, error) {
	return uc.repo.GetReservatorios()
}

func (uc *ReservatorioUseCase) ObterResumoDashboard(reservatorioID int) (*model.DashboardResumo, error) {
	// 1. Busca dados do monitoramento e metas
	ultimoMonit, err := uc.repo.GetUltimoMonitoramento(reservatorioID)
	if err != nil {
		return nil, err
	}

	metas, err := uc.repo.GetMetas(reservatorioID)
	if err != nil {
		return nil, err
	}

	// 2. Cálculo do Estado de Seca (USANDO A NOVA CALCULADORA)
	estadoAtual := uc.calc.CalcularEstado(ultimoMonit, metas)

	// 3. Cálculo de Dias desde a última mudança
	diasDesdeMudanca := 0
	historico, err := uc.repo.GetHistoricoMonitoramento(reservatorioID, 365)
	if err == nil && len(historico) > 1 {
		// (USANDO A NOVA CALCULADORA)
		diasDesdeMudanca = uc.calc.CalcularDiasDesdeMudanca(estadoAtual, historico, metas)
	}

	// ... (Resto do método continua igual: busca planos de ação e monta resumo) ...

	planos, err := uc.repo.GetPlanosAcao(reservatorioID, "", estadoAtual, "", "", "")

	var medidasRecomendadas []model.PlanoAcaoResumo
	if err == nil {
		for _, p := range planos {
			medidasRecomendadas = append(medidasRecomendadas, model.PlanoAcaoResumo{
				Acao:         p.Acoes,
				Descricao:    p.DescricaoAcao,
				Responsaveis: p.Responsaveis,
			})
		}
	} else {
		medidasRecomendadas = []model.PlanoAcaoResumo{}
	}

	resumo := &model.DashboardResumo{
		VolumeAtualHm3:      ultimoMonit.VolumeHm3,
		VolumePercentual:    ultimoMonit.VolumePercentual,
		EstadoAtualSeca:     estadoAtual,
		DataUltimaMedicao:   ultimoMonit.Data.Format("2006-01-02"),
		DiasDesdeMudanca:    diasDesdeMudanca,
		MedidasRecomendadas: medidasRecomendadas,
	}

	return resumo, nil
}

func (uc *ReservatorioUseCase) ListarPlanosAcao(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error) {
	return uc.repo.GetPlanosAcao(reservatorioID, situacao, estado, impacto, problema, acao)
}

func (uc *ReservatorioUseCase) ObterFiltrosPlanoAcao(reservatorioID int) (*model.FiltrosPlanoAcao, error) {
	return uc.repo.GetFiltrosPlanoAcao(reservatorioID)
}

func (uc *ReservatorioUseCase) ListarUsosAgua(reservatorioID int) ([]model.UsoAgua, error) {
	return uc.repo.GetUsosAgua(reservatorioID)
}

func (uc *ReservatorioUseCase) ListarResponsaveis(reservatorioID int) ([]model.Responsavel, error) {
	return uc.repo.GetResponsaveis(reservatorioID)
}

func (uc *ReservatorioUseCase) ObterDadosGrafico(reservatorioID int) ([]model.GraficoVolumeData, error) {
	// 1. Busca TODO o histórico (limit=0) e as Metas
	historico, err := uc.repo.GetHistoricoMonitoramento(reservatorioID, 0)
	if err != nil {
		return nil, err
	}
	metas, err := uc.repo.GetMetas(reservatorioID)
	if err != nil {
		return nil, err
	}

	// 2. Transforma Metas em um Mapa para acesso rápido (O(1))
	// Chave: Mês (int), Valor: Struct Meta
	mapaMetas := make(map[int]model.VolumeMeta)
	for _, m := range metas {
		mapaMetas[m.MesNum] = m
	}

	// 3. Monta a lista de resposta cruzando os dados
	var grafico []model.GraficoVolumeData
	for _, registro := range historico {
		mes := int(registro.Data.Month())
		meta, existe := mapaMetas[mes]

		item := model.GraficoVolumeData{
			Data:   registro.Data.Format("2006-01-02"),
			Volume: registro.VolumeHm3,
		}

		if existe {
			item.Meta1 = meta.Meta1v
			item.Meta2 = meta.Meta2v
			item.Meta3 = meta.Meta3v
		}

		grafico = append(grafico, item)
	}

	return grafico, nil
}

func (uc *ReservatorioUseCase) ObterBalancoHidrico(reservatorioID int) (*model.BalancoHidricoResumo, error) {
	// Busca dados paralelos
	bm, err := uc.repo.GetBalancoMensal(reservatorioID)
	if err != nil {
		return nil, err
	}

	cd, err := uc.repo.GetComposicaoDemanda(reservatorioID)
	if err != nil {
		return nil, err
	}

	od, err := uc.repo.GetOfertaDemanda(reservatorioID)
	if err != nil {
		return nil, err
	}

	// Formatação para JSON igual ao Python
	// Aqui convertemos para maps para formatar nomes de chaves customizados se necessário
	// ou cálculos simples (como Balanço = Afluencia - Demanda)

	var listaBM []map[string]interface{}
	for _, item := range bm {
		listaBM = append(listaBM, map[string]interface{}{
			"Mês":               item.Mes,
			"Afluência (m³/s)":  item.AfluenciaM3s,
			"Demanda (m³/s)":    item.DemandasM3s,
			"Balanço (m³/s)":    item.AfluenciaM3s - item.DemandasM3s, // Cálculo feito on-the-fly
			"Evaporação (m³/s)": item.EvaporacaoM3s,
		})
	}

	var listaCD []map[string]interface{}
	for _, item := range cd {
		listaCD = append(listaCD, map[string]interface{}{
			"Uso":         item.Usos,
			"Vazão (L/s)": item.DemandasHm3,
		})
	}

	var listaOD []map[string]interface{}
	for _, item := range od {
		listaOD = append(listaOD, map[string]interface{}{
			"Cenário":       item.Cenarios,
			"Oferta (L/s)":  item.OfertaM3s,
			"Demanda (L/s)": item.DemandaM3s,
		})
	}

	return &model.BalancoHidricoResumo{
		BalancoMensal:     listaBM,
		ComposicaoDemanda: listaCD,
		OfertaDemanda:     listaOD,
	}, nil
}

func (uc *ReservatorioUseCase) ObterDetalhesReservatorio(id int) (*model.ReservatorioDetalhes, error) {
	res, err := uc.repo.GetReservatorioByID(id)
	if err != nil {
		return nil, err
	}

	// Lógica de URL Base (Igual ao Python)
	baseURL := os.Getenv("RAILWAY_STATIC_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Valor default local
	}

	// Monta as URLs
	var urlImagem, urlUsos string
	if res.NomeImagem != "" {
		urlImagem = baseURL + "/static/images/" + res.NomeImagem
	}
	if res.NomeImagemUsos != "" {
		urlUsos = baseURL + "/static/images/" + res.NomeImagemUsos
	}

	return &model.ReservatorioDetalhes{
		ID:            res.ID,
		Nome:          res.Nome,
		Municipio:     res.Municipio,
		Descricao:     res.Descricao,
		Lat:           res.Lat,
		Long:          res.Long,
		UrlImagem:     urlImagem,
		UrlImagemUsos: urlUsos,
	}, nil
}

func (uc *ReservatorioUseCase) ObterHistoricoTabular(id int) ([]model.HistoricoTabela, error) {
	// 1. Busca TODO o histórico (limit=0)
	historico, err := uc.repo.GetHistoricoMonitoramento(id, 0)
	if err != nil {
		return nil, err
	}

	// 2. Busca metas para cálculo
	metas, err := uc.repo.GetMetas(id)
	if err != nil {
		return nil, err
	}

	// 3. Monta a lista formatada
	var tabela []model.HistoricoTabela
	for _, registro := range historico {
		// (USANDO A NOVA CALCULADORA)
		// Nota: Passamos &registro porque a calculadora espera um ponteiro
		estado := uc.calc.CalcularEstado(&registro, metas)

		tabela = append(tabela, model.HistoricoTabela{
			Data:       registro.Data.Format("02/01/2006"),
			EstadoSeca: estado,
			VolumeHm3:  registro.VolumeHm3,
		})
	}

	return tabela, nil
}

func (uc *ReservatorioUseCase) AtualizarDadosFunceme(reservatorioID int) (int, error) {
	// 1. Busca dados do reservatório para pegar o Código FUNCEME
	res, err := uc.repo.GetReservatorioByID(reservatorioID)
	if err != nil {
		return 0, err
	}
	if res.CodigoFunceme == "" {
		return 0, fmt.Errorf("reservatório sem código FUNCEME cadastrado")
	}

	// 2. Busca dados na API externa usando o serviço dedicado
	// Definimos uma data de início fixa conforme sua lógica original
	novosDados, err := uc.funcemeSvc.BuscarSeriesHistoricas(res.CodigoFunceme, "2023-01-01")
	if err != nil {
		return 0, err
	}

	if len(novosDados) == 0 {
		return 0, nil
	}

	// 3. Busca datas existentes no banco para filtrar duplicatas
	datasExistentes, err := uc.repo.GetDatasMonitoramento(reservatorioID)
	if err != nil {
		return 0, err
	}

	// 4. Filtra e prepara para salvar
	var registrosParaSalvar []model.Monitoramento
	for _, m := range novosDados {
		dataKey := m.Data.Format("2006-01-02")

		// Se não existe no mapa, adiciona
		if !datasExistentes[dataKey] {
			// Atribui o ID do reservatório que o serviço externo desconhece
			m.ReservatorioID = uint(reservatorioID)
			registrosParaSalvar = append(registrosParaSalvar, m)
		}
	}

	// 5. Salva no banco
	if len(registrosParaSalvar) > 0 {
		if err := uc.repo.SalvarMonitoramentos(registrosParaSalvar); err != nil {
			return 0, err
		}
	}

	return len(registrosParaSalvar), nil
}

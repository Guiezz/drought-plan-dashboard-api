package usecase

import (
	"fmt"
	"os"
	"time"

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

	GetDatasMonitoramento(reservatorioID int) (map[string]bool, error)
	SalvarMonitoramentos(registros []model.Monitoramento) error
}

type ReservatorioUseCase struct {
	repo       ReservatorioRepositoryInterface
	planoRepo  PlanoAcaoRepositoryInterface // <--- 2. INJEÇÃO DA NOVA DEPENDÊNCIA
	calc       calculator.SecaCalculator
	funcemeSvc funceme.Service
}

func NewReservatorioUseCase(repo ReservatorioRepositoryInterface, planoRepo PlanoAcaoRepositoryInterface, calc calculator.SecaCalculator, funcemeSvc funceme.Service) *ReservatorioUseCase {
	return &ReservatorioUseCase{
		repo:       repo,
		planoRepo:  planoRepo,
		calc:       calc,
		funcemeSvc: funcemeSvc,
	}
}

func (uc *ReservatorioUseCase) ListarTodos() ([]model.Reservatorio, error) {
	return uc.repo.GetReservatorios()
}

func (uc *ReservatorioUseCase) ListReservoirIDs() ([]uint, error) {
	reservatorios, err := uc.repo.GetReservatorios()
	if err != nil {
		return nil, err
	}
	ids := make([]uint, len(reservatorios))
	for i, r := range reservatorios {
		ids[i] = r.ID
	}
	return ids, nil
}

func (uc *ReservatorioUseCase) ObterResumoDashboard(reservatorioID int) (*model.DashboardResumo, error) {
	ultimoMonit, err := uc.repo.GetUltimoMonitoramento(reservatorioID)
	if err != nil {
		return nil, err
	}

	metas, err := uc.repo.GetMetas(reservatorioID)
	if err != nil {
		return nil, err
	}

	estadoAtual := uc.calc.CalcularEstado(ultimoMonit, metas)

	diasDesdeMudanca := 0
	historico, err := uc.repo.GetHistoricoMonitoramento(reservatorioID, 365)
	if err == nil && len(historico) > 1 {
		diasDesdeMudanca = uc.calc.CalcularDiasDesdeMudanca(estadoAtual, historico, metas)
	}

	// 3. AGORA USA O planoRepo ESPECÍFICO PARA BUSCAR AS AÇÕES
	planos, err := uc.planoRepo.Listar(reservatorioID, "", estadoAtual, "", "", "")

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

func (uc *ReservatorioUseCase) ObterDetalhesReservatorio(id int) (*model.ReservatorioDetalhes, error) {
	res, err := uc.repo.GetReservatorioByID(id)
	if err != nil {
		return nil, err
	}

	// --- CORREÇÃO AQUI ---
	// Tenta pegar a URL do Render, senão usa localhost na porta 8000 (padrão do projeto)
	baseURL := os.Getenv("RENDER_EXTERNAL_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	// Remove barra final para evitar duplicidade (ex: .com//static)
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	// ---------------------

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

func (uc *ReservatorioUseCase) ObterGatilhosPGPS(reservatorioID int) (*model.GatilhosPGPSResponse, error) {
	res, err := uc.repo.GetReservatorioByID(reservatorioID)
	if err != nil {
		return nil, err
	}

	metas, err := uc.repo.GetMetas(reservatorioID)
	if err != nil {
		return nil, err
	}

	gatilhos := make([]model.GatilhoMensalPGPS, len(metas))
	for i, m := range metas {
		gatilhos[i] = model.GatilhoMensalPGPS{
			MesNum:        m.MesNum,
			MesNome:       m.MesNome,
			SecaSeveraHm3: m.Meta1v * res.Capacidadehm3,
			SecaHm3:       m.Meta2v * res.Capacidadehm3,
			AlertaHm3:     m.Meta3v * res.Capacidadehm3,
			NormalHm3:     res.Capacidadehm3,
		}
	}

	return &model.GatilhosPGPSResponse{
		ReservatorioID:   res.ID,
		NomeReservatorio: res.Nome,
		CapacidadeHm3:    res.Capacidadehm3,
		Gatilhos:         gatilhos,
	}, nil
}

func (uc *ReservatorioUseCase) AtualizarDadosFunceme(reservatorioID uint) (int, error) {
	// 1. Busca dados do reservatório para pegar o Código FUNCEME
	res, err := uc.repo.GetReservatorioByID(int(reservatorioID))
	if err != nil {
		return 0, err
	}
	if res.CodigoFunceme == "" {
		return 0, fmt.Errorf("reservatório sem código FUNCEME cadastrado")
	}

	// 2. Busca dados na API externa usando o serviço dedicado
	// Busca apenas o mês atual para reduzir custo computacional
	dataInicio := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	novosDados, err := uc.funcemeSvc.BuscarSeriesHistoricas(res.CodigoFunceme, dataInicio)
	if err != nil {
		return 0, err
	}

	if len(novosDados) == 0 {
		return 0, nil
	}

	// 3. Busca datas existentes no banco para filtrar duplicatas
	datasExistentes, err := uc.repo.GetDatasMonitoramento(int(reservatorioID))
	if err != nil {
		return 0, err
	}

	// 4. Filtra e prepara para salvar
	var registrosParaSalvar []model.Monitoramento
	for _, m := range novosDados {
		dataKey := m.Data.Format("2006-01-02")

		// Se não existe no mapa, adiciona
		if !datasExistentes[dataKey] {
			m.ReservatorioID = reservatorioID
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

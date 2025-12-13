package usecase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

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
	repo ReservatorioRepositoryInterface
}

func NewReservatorioUseCase(repo ReservatorioRepositoryInterface) *ReservatorioUseCase {
	return &ReservatorioUseCase{repo: repo}
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

	// 2. Cálculo do Estado de Seca
	estadoAtual := calcularEstadoSeca(ultimoMonit, metas)

	// 3. Cálculo de Dias desde a última mudança
	diasDesdeMudanca := 0
	historico, err := uc.repo.GetHistoricoMonitoramento(reservatorioID, 365)
	if err == nil && len(historico) > 1 {
		diasDesdeMudanca = calcularDiasDesdeMudanca(estadoAtual, historico, metas)
	}

	// 4. CORREÇÃO: Busca as Medidas Recomendadas baseadas no estado atual
	// A função GetPlanosAcao espera: (id, situacao, estado, impacto, problema, acao)
	// Passamos apenas o 'estado' para filtrar.
	planos, err := uc.repo.GetPlanosAcao(reservatorioID, "", estadoAtual, "", "", "")

	// Mapeia para o DTO simplificado que o Dashboard espera
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
		// Se der erro ao buscar planos, logamos (opcional) mas não quebramos o dashboard inteiro,
		// retornamos lista vazia.
		medidasRecomendadas = []model.PlanoAcaoResumo{}
	}

	// 5. Monta o DTO de resposta final
	resumo := &model.DashboardResumo{
		VolumeAtualHm3:      ultimoMonit.VolumeHm3,
		VolumePercentual:    ultimoMonit.VolumePercentual,
		EstadoAtualSeca:     estadoAtual,
		DataUltimaMedicao:   ultimoMonit.Data.Format("2006-01-02"),
		DiasDesdeMudanca:    diasDesdeMudanca,
		MedidasRecomendadas: medidasRecomendadas, // Agora populado!
	}

	return resumo, nil
}

// --- Funções Auxiliares de Lógica de Negócio ---

func calcularEstadoSeca(m *model.Monitoramento, metas []model.VolumeMeta) string {
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

	// Lógica replicada do Python (np.select)
	// IMPORTANTE: Verifique se no banco 'volume_percentual' é 0-100 ou 0-1.
	// Assumindo 0-100 conforme o seu código Python original parecia tratar na visualização.
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

func calcularDiasDesdeMudanca(estadoAtual string, historico []model.Monitoramento, metas []model.VolumeMeta) int {
	// Itera do mais recente para o mais antigo (excluindo o atual que é o índice 0 ou comparando com ele)
	dataAtual := historico[0].Data

	for _, registro := range historico {
		estadoRegistro := calcularEstadoSeca(&registro, metas)

		// Se o estado for diferente do atual, encontramos a data da mudança
		if estadoRegistro != estadoAtual {
			// Diferença em dias
			diff := dataAtual.Sub(registro.Data)
			return int(diff.Hours() / 24)
		}
	}

	// Se percorreu todo o histórico e não mudou, retorna dias totais do histórico
	diff := dataAtual.Sub(historico[len(historico)-1].Data)
	return int(diff.Hours() / 24)
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

	// Mapa de metas para performance (O(1))
	mapaMetas := make(map[int]model.VolumeMeta)
	for _, m := range metas {
		mapaMetas[m.MesNum] = m
	}

	// 3. Monta a lista formatada
	var tabela []model.HistoricoTabela
	for _, registro := range historico {
		// Calcula estado
		// (Assume-se que calcularEstadoSeca aceita *Monitoramento e []VolumeMeta,
		// mas podemos otimizar passando a meta direta se refatorarmos,
		// por hora usamos a função existente passando o slice completo ou ajustamos)

		// NOTA: Para reusar a função 'calcularEstadoSeca' existente que pede slice:
		estado := calcularEstadoSeca(&registro, metas)

		tabela = append(tabela, model.HistoricoTabela{
			Data:       registro.Data.Format("02/01/2006"), // Formato Brasileiro dd/mm/yyyy
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

	// 2. Prepara a URL (Lógica do Python: de 2023-01-01 até Hoje)
	hoje := time.Now().Format("2006-01-02")
	url := fmt.Sprintf("https://apil5.funceme.br/rpc/v1/reservatorio-series?reservatorio_id=%s&data_inicio=2023-01-01&data_fim=%s", res.CodigoFunceme, hoje)

	// 3. Faz a requisição HTTP GET
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("erro ao conectar na FUNCEME: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("API FUNCEME retornou status: %d", resp.StatusCode)
	}

	// 4. Decodifica o JSON
	var funcemeResp model.FuncemeResponse
	if err := json.NewDecoder(resp.Body).Decode(&funcemeResp); err != nil {
		return 0, fmt.Errorf("erro ao ler JSON da FUNCEME: %v", err)
	}

	listaDados := funcemeResp.Data.List
	if len(listaDados) == 0 {
		return 0, nil // Nenhum dado novo
	}

	// 5. Busca datas existentes no banco para filtrar duplicatas
	datasExistentes, err := uc.repo.GetDatasMonitoramento(reservatorioID)
	if err != nil {
		return 0, err
	}

	// 6. Filtra e cria objetos para salvar
	var novosRegistros []model.Monitoramento
	for _, item := range listaDados {
		// A data vem "2023-10-25 00:00:00" ou similar, pegamos só a data
		dataSimples := item.DataStr
		if len(dataSimples) >= 10 {
			dataSimples = dataSimples[:10]
		}

		// Se não existe no mapa, adiciona
		if !datasExistentes[dataSimples] {
			// Parse da string para time.Time do Go
			dataTime, _ := time.Parse("2006-01-02", dataSimples)

			novosRegistros = append(novosRegistros, model.Monitoramento{
				ReservatorioID:   uint(reservatorioID),
				Data:             dataTime,
				VolumeHm3:        item.Volume,
				VolumePercentual: item.VolumePerc,
			})
		}
	}

	// 7. Salva no banco
	if len(novosRegistros) > 0 {
		if err := uc.repo.SalvarMonitoramentos(novosRegistros); err != nil {
			return 0, err
		}
	}

	return len(novosRegistros), nil
}

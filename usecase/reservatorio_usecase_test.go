package usecase

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockReservatorioRepo struct {
	mock.Mock
}

func (m *mockReservatorioRepo) GetReservatorios() ([]model.Reservatorio, error) {
	args := m.Called()
	return args.Get(0).([]model.Reservatorio), args.Error(1)
}

func (m *mockReservatorioRepo) GetReservatorioByID(id int) (*model.Reservatorio, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Reservatorio), args.Error(1)
}

func (m *mockReservatorioRepo) GetUltimoMonitoramento(reservatorioID int) (*model.Monitoramento, error) {
	args := m.Called(reservatorioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Monitoramento), args.Error(1)
}

func (m *mockReservatorioRepo) GetMetas(reservatorioID int) ([]model.VolumeMeta, error) {
	args := m.Called(reservatorioID)
	return args.Get(0).([]model.VolumeMeta), args.Error(1)
}

func (m *mockReservatorioRepo) GetHistoricoMonitoramento(reservatorioID int, limit int) ([]model.Monitoramento, error) {
	args := m.Called(reservatorioID, limit)
	return args.Get(0).([]model.Monitoramento), args.Error(1)
}

func (m *mockReservatorioRepo) GetDatasMonitoramento(reservatorioID int) (map[string]bool, error) {
	args := m.Called(reservatorioID)
	return args.Get(0).(map[string]bool), args.Error(1)
}

func (m *mockReservatorioRepo) SalvarMonitoramentos(registros []model.Monitoramento) error {
	args := m.Called(registros)
	return args.Error(0)
}

type mockPlanoAcaoRepo struct {
	mock.Mock
}

func (m *mockPlanoAcaoRepo) Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error) {
	args := m.Called(reservatorioID, situacao, estado, impacto, problema, acao)
	return args.Get(0).([]model.PlanoAcao), args.Error(1)
}

func (m *mockPlanoAcaoRepo) ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error) {
	args := m.Called(reservatorioID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.FiltrosPlanoAcao), args.Error(1)
}

func (m *mockPlanoAcaoRepo) AtualizarStatus(acaoID, usuarioID uint, novaSituacao string) error {
	args := m.Called(acaoID, usuarioID, novaSituacao)
	return args.Error(0)
}

type mockSecaCalculator struct {
	mock.Mock
}

func (m *mockSecaCalculator) CalcularEstado(monit *model.Monitoramento, metas []model.VolumeMeta) string {
	args := m.Called(monit, metas)
	return args.String(0)
}

func (m *mockSecaCalculator) CalcularDiasDesdeMudanca(estadoAtual string, historico []model.Monitoramento, metas []model.VolumeMeta) int {
	args := m.Called(estadoAtual, historico, metas)
	return args.Int(0)
}

func TestObterResumoDashboard(t *testing.T) {
	mockRepo := new(mockReservatorioRepo)
	mockPlano := new(mockPlanoAcaoRepo)
	mockCalc := new(mockSecaCalculator)

	uc := NewReservatorioUseCase(mockRepo, mockPlano, mockCalc, nil)

	ultimoMonit := &model.Monitoramento{
		Data:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		VolumeHm3:        50,
		VolumePercentual: 50,
	}

	metas := []model.VolumeMeta{{MesNum: 1, Meta1v: 0.2, Meta2v: 0.4, Meta3v: 0.6}}

	t.Run("resumo completo com ações recomendadas", func(t *testing.T) {
		mockRepo.On("GetUltimoMonitoramento", 1).Return(ultimoMonit, nil).Once()
		mockRepo.On("GetMetas", 1).Return(metas, nil).Once()
		mockCalc.On("CalcularEstado", ultimoMonit, metas).Return("ALERTA").Once()
		historicoComDois := []model.Monitoramento{
			{Data: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), VolumePercentual: 50},
			{Data: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), VolumePercentual: 30},
		}
		mockRepo.On("GetHistoricoMonitoramento", 1, 365).Return(historicoComDois, nil).Once()
		mockCalc.On("CalcularDiasDesdeMudanca", "ALERTA", historicoComDois, metas).Return(14).Once()
		mockPlano.On("Listar", 1, "", "ALERTA", "", "", "").Return([]model.PlanoAcao{
			{Acoes: "Ação 1", DescricaoAcao: "Desc 1", Responsaveis: "Resp 1"},
		}, nil).Once()

		resumo, err := uc.ObterResumoDashboard(1)

		assert.NoError(t, err)
		assert.Equal(t, 50.0, resumo.VolumeAtualHm3)
		assert.Equal(t, "ALERTA", resumo.EstadoAtualSeca)
		assert.Len(t, resumo.MedidasRecomendadas, 1)
		mockRepo.AssertExpectations(t)
		mockPlano.AssertExpectations(t)
		mockCalc.AssertExpectations(t)
	})

	t.Run("erro ao buscar monitoramento", func(t *testing.T) {
		mockRepo.On("GetUltimoMonitoramento", 2).Return(nil, errors.New("not found")).Once()

		resumo, err := uc.ObterResumoDashboard(2)

		assert.Error(t, err)
		assert.Nil(t, resumo)
	})
}

func TestListarTodos(t *testing.T) {
	mockRepo := new(mockReservatorioRepo)
	uc := NewReservatorioUseCase(mockRepo, nil, nil, nil)

	mockRepo.On("GetReservatorios").Return([]model.Reservatorio{
		{Nome: "Açude 1"},
		{Nome: "Açude 2"},
	}, nil).Once()

	reservatorios, err := uc.ListarTodos()
	assert.NoError(t, err)
	assert.Len(t, reservatorios, 2)
	assert.Equal(t, "Açude 1", reservatorios[0].Nome)
	mockRepo.AssertExpectations(t)
}

func TestListReservoirIDs(t *testing.T) {
	mockRepo := new(mockReservatorioRepo)
	uc := NewReservatorioUseCase(mockRepo, nil, nil, nil)

	mockRepo.On("GetReservatorios").Return([]model.Reservatorio{
		{ID: 1, Nome: "A"},
		{ID: 2, Nome: "B"},
	}, nil).Once()

	ids, err := uc.ListReservoirIDs()
	assert.NoError(t, err)
	assert.Equal(t, []uint{1, 2}, ids)
	mockRepo.AssertExpectations(t)
}

func TestObterDetalhesReservatorio_FichaTecnica(t *testing.T) {
	mockRepo := new(mockReservatorioRepo)
	uc := NewReservatorioUseCase(mockRepo, new(mockPlanoAcaoRepo), new(mockSecaCalculator), nil)

	area := 206.2
	ano := 1924
	construtor := "IFOCS (atual DNOCS)"
	mockRepo.On("GetReservatorioByID", 1).Return(&model.Reservatorio{
		ID:              1,
		Nome:            "Acarape do Meio",
		Capacidadehm3:   29.6,
		CodigoFunceme:   "78",
		AreaDrenagemKm2: &area,
		AnoConstrucao:   &ano,
		Construtor:      &construtor,
		RioBarrado:      nil,
	}, nil).Once()

	det, err := uc.ObterDetalhesReservatorio(1)

	assert.NoError(t, err)
	assert.Equal(t, 29.6, det.CapacidadeHm3)
	assert.Equal(t, "78", det.CodigoFunceme)
	assert.Equal(t, 206.2, *det.AreaDrenagemKm2)
	assert.Equal(t, 1924, *det.AnoConstrucao)
	assert.Equal(t, "IFOCS (atual DNOCS)", *det.Construtor)
	assert.Nil(t, det.RioBarrado)

	raw, err := json.Marshal(det)
	assert.NoError(t, err)
	assert.NotContains(t, string(raw), "rio_barrado")
	assert.Contains(t, string(raw), `"area_drenagem_km2":206.2`)
	mockRepo.AssertExpectations(t)
}

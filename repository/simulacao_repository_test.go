package repository

import (
	"testing"

	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func seedSimulacaoData(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Create(&simulador.SimAcude{
		Codigo:        1,
		Nome:          "Açude Sim",
		Capacidade:    100_000_000,
		Municipio:     "Quixadá",
		EstacaoEvapID: 10,
	})
	db.Create(&simulador.SimCAV{AcudeID: 1, Cota: 100, Area: 10, Volume: 0})
	db.Create(&simulador.SimCAV{AcudeID: 1, Cota: 105, Area: 15, Volume: 50})
	db.Create(&simulador.SimCAV{AcudeID: 1, Cota: 110, Area: 20, Volume: 100})
	db.Create(&simulador.SimEvaporacao{
		Codigo: 10, Jan: 100, Fev: 90, Mar: 80, Abr: 70, Mai: 60, Jun: 50,
		Jul: 50, Ago: 60, Set: 70, Out: 80, Nov: 90, Dez: 100,
	})
	db.Create(&simulador.SimVazao{AcudeID: 1, Ano: 2020, Jan: 5, Fev: 6, Mar: 7, Abr: 4, Mai: 3, Jun: 2,
		Jul: 1, Ago: 1, Set: 2, Out: 3, Nov: 4, Dez: 5})
	db.Create(&simulador.SimVazao{AcudeID: 1, Ano: 2021, Jan: 10, Fev: 11, Mar: 12, Abr: 8, Mai: 6, Jun: 4,
		Jul: 3, Ago: 3, Set: 4, Out: 5, Nov: 6, Dez: 8})
}

func TestSimulacaoRepository_GetAcude(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	acude, err := repo.GetAcude(1)
	assert.NoError(t, err)
	assert.Equal(t, "Açude Sim", acude.Nome)
	assert.Equal(t, 10, acude.EstacaoEvapID)

	_, err = repo.GetAcude(999)
	assert.Error(t, err)
}

func TestSimulacaoRepository_GetCAV(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	cav, err := repo.GetCAV(1)
	assert.NoError(t, err)
	assert.Len(t, cav, 3)
	assert.True(t, cav[0].Volume < cav[2].Volume)
}

func TestSimulacaoRepository_GetEvaporacao(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	evap, err := repo.GetEvaporacao(10)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, evap.Jan)
	assert.Equal(t, 50.0, evap.Jun)
}

func TestSimulacaoRepository_GetVazoes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	vazoes, err := repo.GetVazoes(1, 2020, 2021)
	assert.NoError(t, err)
	assert.Len(t, vazoes, 2)

	vazoes, err = repo.GetVazoes(1, 2020, 2020)
	assert.NoError(t, err)
	assert.Len(t, vazoes, 1)
}

func TestSimulacaoRepository_GetVazoesByAnos(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	vazoes, err := repo.GetVazoesByAnos(1, []int{2020})
	assert.NoError(t, err)
	assert.Len(t, vazoes, 1)
	assert.Equal(t, 2020, vazoes[0].Ano)

	vazoes, err = repo.GetVazoesByAnos(1, []int{2020, 2021})
	assert.NoError(t, err)
	assert.Len(t, vazoes, 2)
}

func TestSimulacaoRepository_GetAnosVazoes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	anos, err := repo.GetAnosVazoes(1)
	assert.NoError(t, err)
	assert.Equal(t, []int{2020, 2021}, anos)
}

func TestSimulacaoRepository_ListarAcudes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSimulacaoRepository(db)
	seedSimulacaoData(t, db)

	acudes, err := repo.ListarAcudes()
	assert.NoError(t, err)
	assert.Len(t, acudes, 1)
	assert.Equal(t, "Açude Sim", acudes[0].Nome)
}

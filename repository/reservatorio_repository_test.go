package repository

import (
	"testing"
	"time"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedReservatorio(t *testing.T, db *gorm.DB) {
	t.Helper()
	res := model.Reservatorio{Nome: "Açude Teste", Capacidadehm3: 100, CodigoFunceme: "123"}
	require.NoError(t, db.Create(&res).Error)
}

func seedMonitoramento(t *testing.T, db *gorm.DB, reservatorioID uint, dates []string, volumes []float64) {
	t.Helper()
	for i, d := range dates {
		data, _ := time.Parse("2006-01-02", d)
		require.NoError(t, db.Create(&model.Monitoramento{
			ReservatorioID:   reservatorioID,
			Data:             data,
			VolumeHm3:        volumes[i],
			VolumePercentual: volumes[i],
		}).Error)
	}
}

func TestReservatorioRepository_GetReservatorios(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)

	t.Run("retorna lista vazia quando não há reservatórios", func(t *testing.T) {
		res, err := repo.GetReservatorios()
		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("retorna ordenado por nome", func(t *testing.T) {
		db.Create(&model.Reservatorio{Nome: "Zé"})
		db.Create(&model.Reservatorio{Nome: "Beta"})
		db.Create(&model.Reservatorio{Nome: "Alfa"})

		res, err := repo.GetReservatorios()
		assert.NoError(t, err)
		assert.Len(t, res, 3)
		assert.Equal(t, "Alfa", res[0].Nome)
		assert.Equal(t, "Beta", res[1].Nome)
		assert.Equal(t, "Zé", res[2].Nome)
	})
}

func TestReservatorioRepository_GetUltimoMonitoramento(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)
	seedReservatorio(t, db)

	t.Run("retorna o registro mais recente", func(t *testing.T) {
		res := model.Reservatorio{}
		db.First(&res)

		seedMonitoramento(t, db, res.ID,
			[]string{"2024-01-01", "2024-06-01", "2024-03-01"},
			[]float64{50, 80, 60},
		)

		monit, err := repo.GetUltimoMonitoramento(int(res.ID))
		assert.NoError(t, err)
		assert.Equal(t, 80.0, monit.VolumeHm3)
	})

	t.Run("erro quando não há monitoramento", func(t *testing.T) {
		_, err := repo.GetUltimoMonitoramento(999)
		assert.Error(t, err)
	})
}

func TestReservatorioRepository_GetHistoricoMonitoramento(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)
	seedReservatorio(t, db)

	res := model.Reservatorio{}
	db.First(&res)
	seedMonitoramento(t, db, res.ID,
		[]string{"2024-01-01", "2024-02-01", "2024-03-01"},
		[]float64{10, 20, 30},
	)

	t.Run("sem limit retorna tudo ASC", func(t *testing.T) {
		hist, err := repo.GetHistoricoMonitoramento(int(res.ID), 0)
		assert.NoError(t, err)
		assert.Len(t, hist, 3)
		assert.Equal(t, 10.0, hist[0].VolumeHm3)
		assert.Equal(t, 30.0, hist[2].VolumeHm3)
	})

	t.Run("com limit retorna N registros DESC", func(t *testing.T) {
		hist, err := repo.GetHistoricoMonitoramento(int(res.ID), 2)
		assert.NoError(t, err)
		assert.Len(t, hist, 2)
		assert.Equal(t, 30.0, hist[0].VolumeHm3)
		assert.Equal(t, 20.0, hist[1].VolumeHm3)
	})
}

func TestReservatorioRepository_GetDatasMonitoramento(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)
	seedReservatorio(t, db)

	res := model.Reservatorio{}
	db.First(&res)

	seedMonitoramento(t, db, res.ID,
		[]string{"2024-01-01", "2024-01-15", "2024-02-01"},
		[]float64{10, 15, 20},
	)

	datas, err := repo.GetDatasMonitoramento(int(res.ID))
	assert.NoError(t, err)
	assert.Len(t, datas, 3)
	assert.True(t, datas["2024-01-01"])
	assert.True(t, datas["2024-02-01"])
}

func TestReservatorioRepository_SalvarMonitoramentos(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)
	seedReservatorio(t, db)

	res := model.Reservatorio{}
	db.First(&res)

	registros := []model.Monitoramento{
		{ReservatorioID: res.ID, Data: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), VolumeHm3: 50, VolumePercentual: 50},
		{ReservatorioID: res.ID, Data: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), VolumeHm3: 40, VolumePercentual: 40},
	}

	err := repo.SalvarMonitoramentos(registros)
	assert.NoError(t, err)

	var count int64
	db.Model(&model.Monitoramento{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestReservatorioRepository_GetReservatorioByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewReservatorioRepository(db)
	seedReservatorio(t, db)

	res, err := repo.GetReservatorioByID(1)
	assert.NoError(t, err)
	assert.Equal(t, "Açude Teste", res.Nome)

	_, err = repo.GetReservatorioByID(999)
	assert.Error(t, err)
}

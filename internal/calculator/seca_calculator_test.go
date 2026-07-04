package calculator

import (
	"testing"
	"time"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
)

func mesMeta(mesNum int, meta1v, meta2v, meta3v float64) model.VolumeMeta {
	return model.VolumeMeta{
		MesNum:  mesNum,
		MesNome: time.Month(mesNum).String()[:3],
		Meta1v:  meta1v,
		Meta2v:  meta2v,
		Meta3v:  meta3v,
	}
}

func TestCalcularEstado(t *testing.T) {
	calc := NewSecaCalculator()

	metas := []model.VolumeMeta{
		mesMeta(1, 0.20, 0.40, 0.60),
		mesMeta(2, 0.25, 0.45, 0.65),
	}

	t.Run("seca severa quando volume abaixo de meta1v", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 15.0,
		}
		assert.Equal(t, "SECA SEVERA", calc.CalcularEstado(m, metas))
	})

	t.Run("seca quando volume entre meta1v e meta2v", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 30.0,
		}
		assert.Equal(t, "SECA", calc.CalcularEstado(m, metas))
	})

	t.Run("alerta quando volume entre meta2v e meta3v", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 50.0,
		}
		assert.Equal(t, "ALERTA", calc.CalcularEstado(m, metas))
	})

	t.Run("normal quando volume acima de meta3v", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 70.0,
		}
		assert.Equal(t, "NORMAL", calc.CalcularEstado(m, metas))
	})

	t.Run("não definido quando monitoramento é nil", func(t *testing.T) {
		assert.Equal(t, "NÃO DEFINIDO", calc.CalcularEstado(nil, metas))
	})

	t.Run("não definido quando mês não tem meta", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 50.0,
		}
		assert.Equal(t, "NÃO DEFINIDO", calc.CalcularEstado(m, metas))
	})

	t.Run("exato no limite de meta1v deve ser seca (comparação estrita <)", func(t *testing.T) {
		m := &model.Monitoramento{
			Data:             time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
			VolumePercentual: 25.0,
		}
		assert.Equal(t, "SECA", calc.CalcularEstado(m, metas))
	})
}

func TestCalcularDiasDesdeMudanca(t *testing.T) {
	calc := NewSecaCalculator()
	metas := []model.VolumeMeta{mesMeta(1, 0.20, 0.40, 0.60)}

	t.Run("zero quando histórico vazio", func(t *testing.T) {
		assert.Equal(t, 0, calc.CalcularDiasDesdeMudanca("NORMAL", nil, metas))
	})

	t.Run("zero quando lista vazia", func(t *testing.T) {
		assert.Equal(t, 0, calc.CalcularDiasDesdeMudanca("NORMAL", []model.Monitoramento{}, metas))
	})

	t.Run("span total quando nunca mudou", func(t *testing.T) {
		historico := []model.Monitoramento{
			{Data: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), VolumePercentual: 70},
			{Data: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), VolumePercentual: 70},
		}
		dias := calc.CalcularDiasDesdeMudanca("NORMAL", historico, metas)
		assert.Equal(t, 30, dias)
	})

	t.Run("retorna dias desde a mudança no meio do histórico", func(t *testing.T) {
		historico := []model.Monitoramento{
			{Data: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC), VolumePercentual: 70},
			{Data: time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC), VolumePercentual: 70},
			{Data: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC), VolumePercentual: 30},
		}
		dias := calc.CalcularDiasDesdeMudanca("NORMAL", historico, metas)
		assert.Equal(t, 21, dias)
	})
}

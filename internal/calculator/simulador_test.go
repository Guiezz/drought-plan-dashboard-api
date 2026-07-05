package calculator

import (
	"testing"

	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/stretchr/testify/assert"
)

func TestFitPolynomial3(t *testing.T) {
	calc := &SimuladorHidrico{}

	t.Run("pontos de uma curva cúbica perfeita", func(t *testing.T) {
		x := []float64{0, 1, 2, 3}
		y := []float64{1, 3, 17, 55}
		coeffs, err := calc.FitPolynomial3(x, y)
		assert.NoError(t, err)
		assert.Len(t, coeffs, 4)
	})

	t.Run("3 pontos também funciona (mínimo necessário para grau 3 seriam 4)", func(t *testing.T) {
		x := []float64{0, 1, 2}
		y := []float64{0, 1, 8}
		_, err := calc.FitPolynomial3(x, y)
		assert.NoError(t, err)
	})
}

func TestEvalPolynomial(t *testing.T) {
	calc := &SimuladorHidrico{}

	t.Run("x=0 retorna coeff[3]", func(t *testing.T) {
		coeffs := []float64{2, 3, 4, 5}
		result := calc.EvalPolynomial(coeffs, 0)
		assert.Equal(t, 5.0, result)
	})

	t.Run("x=1 retorna soma de todos coeficientes", func(t *testing.T) {
		coeffs := []float64{2, 3, 4, 5}
		result := calc.EvalPolynomial(coeffs, 1)
		assert.Equal(t, 14.0, result)
	})
}

func TestSimular(t *testing.T) {
	calc := &SimuladorHidrico{}

	cav := []simulador.SimCAV{
		{Volume: 0, Area: 0},
		{Volume: 50, Area: 10},
		{Volume: 100, Area: 20},
		{Volume: 150, Area: 30},
	}

	t.Run("volume inicial maior que capacidade deve verter", func(t *testing.T) {
		resultado := calc.Simular(
			200,
			cav,
			[]float64{10, 10},
			[]float64{5, 5},
			[]float64{10, 10},
			100,
		)
		assert.Len(t, resultado.Resultados, 2)
		assert.Equal(t, 100.0, resultado.Resultados[0].VolumeFinal)
		assert.True(t, resultado.Resultados[0].Vertimento > 0)
	})

	t.Run("sem afluencia evaporacao zero demanda zero", func(t *testing.T) {
		resultado := calc.Simular(
			50,
			cav,
			[]float64{0, 0, 0},
			[]float64{0, 0, 0},
			[]float64{0, 0, 0},
			100,
		)
		assert.Len(t, resultado.Resultados, 3)
		for _, r := range resultado.Resultados {
			assert.Equal(t, 50.0, r.VolumeFinal)
		}
	})

	t.Run("demanda maior que saldo registra falha", func(t *testing.T) {
		resultado := calc.Simular(
			10,
			cav,
			[]float64{0, 0},
			[]float64{100, 100},
			[]float64{0, 0},
			100,
		)
		assert.True(t, resultado.FrequenciaNaoAtendida > 0)
	})

	t.Run("afluencia enche o reservatorio", func(t *testing.T) {
		resultado := calc.Simular(
			10,
			cav,
			[]float64{200, 200},
			[]float64{5, 5},
			[]float64{0, 0},
			100,
		)
		assert.Equal(t, 100.0, resultado.Resultados[0].VolumeFinal)
		assert.True(t, resultado.Resultados[0].Vertimento > 0)
	})
}

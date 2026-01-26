package calculator

import (
	"math"

	"github.com/guiezz/dashboard-api/model/simulador"
	"gonum.org/v1/gonum/mat"
)

type SimuladorHidrico struct{}

func NewSimuladorHidrico() *SimuladorHidrico {
	return &SimuladorHidrico{}
}

// FitPolynomial3 calcula os coeficientes de um polinômio de grau 3 (ax³ + bx² + cx + d)
func (s *SimuladorHidrico) FitPolynomial3(x, y []float64) ([]float64, error) {
	rows := len(x)
	degree := 4 // grau 3 tem 4 coeficientes

	A := mat.NewDense(rows, degree, nil)
	Y := mat.NewDense(rows, 1, y)

	for i, v := range x {
		A.Set(i, 0, math.Pow(v, 3))
		A.Set(i, 1, math.Pow(v, 2))
		A.Set(i, 2, v)
		A.Set(i, 3, 1)
	}

	var xHat mat.Dense
	err := xHat.Solve(A, Y)
	if err != nil {
		return nil, err
	}

	coeffs := make([]float64, degree)
	for i := 0; i < degree; i++ {
		coeffs[i] = xHat.At(i, 0)
	}
	return coeffs, nil
}

// EvalPolynomial calcula o valor y para um dado x usando os coeficientes
func (s *SimuladorHidrico) EvalPolynomial(coeffs []float64, x float64) float64 {
	return coeffs[0]*math.Pow(x, 3) +
		coeffs[1]*math.Pow(x, 2) +
		coeffs[2]*x +
		coeffs[3]
}

// Simular executa o balanço hídrico
func (s *SimuladorHidrico) Simular(
	volInicial float64,
	cav []simulador.SimCAV,
	afluencias []float64,
	demandas []float64,
	evaporacao []float64,
	volMax float64,
) simulador.SimulacaoResponse {

	// 1. Prepara dados para regressão (Volume -> Área)
	var xVol, yArea []float64
	for _, p := range cav {
		xVol = append(xVol, p.Volume)
		yArea = append(yArea, p.Area)
	}

	// 2. Calcula polinômio (Volume -> Área)
	coeffs, err := s.FitPolynomial3(xVol, yArea)
	if err != nil {
		coeffs = []float64{0, 0, 0, 0}
	}

	nMeses := len(afluencias)
	resultados := make([]simulador.SimulacaoResultado, nMeses)
	volumeAtual := volInicial
	mesesNaoAtendidos := 0

	for t := 0; t < nMeses; t++ {
		// Armazena volume INICIAL para exibição
		volInicioMes := volumeAtual

		// a. Calcula Área (km² -> m²)
		areaKm2 := s.EvalPolynomial(coeffs, volumeAtual)
		if areaKm2 < 0 {
			areaKm2 = 0
		}

		// Trava de segurança (1000 km²)
		if areaKm2 > 1000 {
			areaKm2 = 1000
		}

		areaM2 := areaKm2 * 1e6

		// b. Calcula Evaporação
		evapM := evaporacao[t] / 1000.0
		evapHm3 := (areaM2 * evapM) / 1e6

		// c. Balanço
		afluencia := afluencias[t]
		demanda := demandas[t]
		volMorto := 0.0

		saldo := volumeAtual + afluencia - evapHm3 - volMorto
		if saldo < 0 {
			saldo = 0
		}

		retiradaReal := math.Min(demanda, saldo)

		// Verifica falha (Considerando uma margem de erro mínima para float)
		if retiradaReal < (demanda - 0.00001) {
			mesesNaoAtendidos++
		}

		vPotencial := volumeAtual + afluencia - evapHm3 - retiradaReal
		if vPotencial < 0 {
			vPotencial = 0
		}

		vertimento := 0.0
		if vPotencial > volMax {
			vertimento = vPotencial - volMax
			vPotencial = volMax
		}

		volumeAtual = vPotencial

		// AQUI ESTAVA O ERRO: Removido math.Round para manter precisão
		// O Frontend que decida quantas casas mostrar
		resultados[t] = simulador.SimulacaoResultado{
			Data:       "",
			Volume:     volInicioMes,
			Afluencia:  afluencia,
			Retirada:   retiradaReal,
			Evaporacao: evapHm3,
			Vertimento: vertimento,
		}
	}

	freq := (float64(mesesNaoAtendidos) / float64(nMeses)) * 100

	return simulador.SimulacaoResponse{
		Resultados:            resultados,
		FrequenciaNaoAtendida: freq,
		VolumeFinal:           volumeAtual,
	}
}

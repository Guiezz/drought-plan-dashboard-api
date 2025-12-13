package funceme

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/guiezz/dashboard-api/model"
)

// Service define o contrato para buscar dados da Funceme
type Service interface {
	BuscarSeriesHistoricas(codigoFunceme string, dataInicio string) ([]model.Monitoramento, error)
}

type funcemeService struct {
	client *http.Client
}

// NewFuncemeService cria uma nova instância do serviço com timeout configurado
func NewFuncemeService() Service {
	return &funcemeService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *funcemeService) BuscarSeriesHistoricas(codigoFunceme string, dataInicio string) ([]model.Monitoramento, error) {
	// Define data fim como hoje
	hoje := time.Now().Format("2006-01-02")

	// Monta a URL
	url := fmt.Sprintf("https://apil5.funceme.br/rpc/v1/reservatorio-series?reservatorio_id=%s&data_inicio=%s&data_fim=%s",
		codigoFunceme, dataInicio, hoje)

	// Faz a requisição
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar na FUNCEME: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API FUNCEME retornou status: %d", resp.StatusCode)
	}

	// Decodifica o JSON
	var funcemeResp model.FuncemeResponse
	if err := json.NewDecoder(resp.Body).Decode(&funcemeResp); err != nil {
		return nil, fmt.Errorf("erro ao ler JSON da FUNCEME: %v", err)
	}

	var monitoramentos []model.Monitoramento

	// Converte os dados brutos para o modelo de domínio
	for _, item := range funcemeResp.Data.List {
		// A data vem como string, precisamos converter
		dataStr := item.DataStr
		if len(dataStr) >= 10 {
			dataStr = dataStr[:10]
		}

		dataTime, err := time.Parse("2006-01-02", dataStr)
		if err != nil {
			continue // Pula registros com data inválida
		}

		// Cria o objeto Monitoramento (sem ID do reservatório ainda)
		monitoramentos = append(monitoramentos, model.Monitoramento{
			Data:             dataTime,
			VolumeHm3:        item.Volume,
			VolumePercentual: item.VolumePerc,
		})
	}

	return monitoramentos, nil
}

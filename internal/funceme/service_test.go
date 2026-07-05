package funceme

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
)

func TestBuscarSeriesHistoricas(t *testing.T) {
	t.Run("sucesso com dados válidos", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.String(), "reservatorio_id=123")
			assert.Contains(t, r.URL.String(), "data_inicio=2024-01-01")

			resp := model.FuncemeResponse{
				Data: struct {
					List []model.FuncemeRegistro `json:"list"`
				}{
					List: []model.FuncemeRegistro{
						{DataStr: "2024-01-15", Volume: 100, VolumePerc: 50},
						{DataStr: "2024-02-15", Volume: 90, VolumePerc: 45},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		svc := NewFuncemeService(server.URL)
		resultado, err := svc.BuscarSeriesHistoricas("123", "2024-01-01")

		assert.NoError(t, err)
		assert.Len(t, resultado, 2)
		assert.Equal(t, 100.0, resultado[0].VolumeHm3)
		assert.Equal(t, 45.0, resultado[1].VolumePercentual)
	})

	t.Run("API retorna erro 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := NewFuncemeService(server.URL)
		_, err := svc.BuscarSeriesHistoricas("123", "2024-01-01")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status: 500")
	})

	t.Run("API retorna dados vazios", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := model.FuncemeResponse{
				Data: struct {
					List []model.FuncemeRegistro `json:"list"`
				}{List: []model.FuncemeRegistro{}},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		svc := NewFuncemeService(server.URL)
		resultado, err := svc.BuscarSeriesHistoricas("123", "2024-01-01")

		assert.NoError(t, err)
		assert.Empty(t, resultado)
	})
}

package repository

import (
	"testing"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanoAcaoRepository_AtualizarStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlanoAcaoRepository(db)

	usuario := model.Usuario{Nome: "Admin", Email: "admin@test.com", SenhaHash: "hash", Role: "cogerh"}
	require.NoError(t, db.Create(&usuario).Error)

	plano := model.PlanoAcao{
		ReservatorioID: 1,
		EstadoSeca:     "ALERTA",
		Acoes:          "Ação teste",
		Situacao:       "Não iniciado",
	}
	require.NoError(t, db.Create(&plano).Error)
	planoID := plano.ID

	t.Run("atualiza status e insere historico na transacao", func(t *testing.T) {
		err := repo.AtualizarStatus(planoID, usuario.ID, "Em andamento")
		assert.NoError(t, err)

		var atualizado model.PlanoAcao
		db.First(&atualizado, planoID)
		assert.Equal(t, "Em andamento", atualizado.Situacao)

		var historico model.HistoricoAcao
		err = db.Where("plano_acao_id = ?", planoID).First(&historico).Error
		assert.NoError(t, err)
		assert.Equal(t, "Não iniciado", historico.SituacaoAnterior)
		assert.Equal(t, "Em andamento", historico.SituacaoNova)
		assert.Equal(t, usuario.ID, historico.UsuarioID)
	})

	t.Run("erro ao buscar acao inexistente", func(t *testing.T) {
		err := repo.AtualizarStatus(9999, usuario.ID, "Concluído")
		assert.Error(t, err)
	})
}

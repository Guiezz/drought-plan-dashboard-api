package repository

import (
	"time"

	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

type PlanoAcaoRepository struct {
	db *gorm.DB
}

func NewPlanoAcaoRepository(db *gorm.DB) *PlanoAcaoRepository {
	return &PlanoAcaoRepository{db: db}
}

func (r *PlanoAcaoRepository) Listar(reservatorioID int, situacao, estado, impacto, problema, acao string) ([]model.PlanoAcao, error) {
	planos := make([]model.PlanoAcao, 0)

	query := r.db.Where("reservatorio_id = ?", reservatorioID)

	if situacao != "" {
		query = query.Where("situacao = ?", situacao)
	}
	if estado != "" {
		query = query.Where("estado_seca = ?", estado)
	}
	if impacto != "" {
		query = query.Where("tipos_impactos = ?", impacto)
	}
	if problema != "" {
		query = query.Where("problemas = ?", problema)
	}
	if acao != "" {
		query = query.Where("acoes = ?", acao)
	}

	result := query.Find(&planos)
	return planos, result.Error
}

func (r *PlanoAcaoRepository) AtualizarStatus(acaoID uint, usuarioID uint, novaSituacao string) error {
	// Inicia uma transação no banco de dados
	return r.db.Transaction(func(tx *gorm.DB) error {
		var plano model.PlanoAcao

		// 1. Busca a ação atual para descobrir a Situação Anterior
		if err := tx.First(&plano, acaoID).Error; err != nil {
			return err
		}

		situacaoAnterior := plano.Situacao

		// 2. Atualiza o status e a autoria no Plano de Ação
		if err := tx.Model(&plano).Updates(map[string]interface{}{
			"situacao":          novaSituacao,
			"atualizado_por_id": usuarioID,
			"updated_at":        time.Now(),
		}).Error; err != nil {
			return err
		}

		// 3. Cria o registro na tabela de histórico
		historico := model.HistoricoAcao{
			PlanoAcaoID:      acaoID,
			UsuarioID:        usuarioID,
			SituacaoAnterior: situacaoAnterior,
			SituacaoNova:     novaSituacao,
			DataAlteracao:    time.Now(),
		}

		if err := tx.Create(&historico).Error; err != nil {
			return err
		}

		// Retornar nil confirma a transação
		return nil
	})
}

func (r *PlanoAcaoRepository) ObterFiltros(reservatorioID int) (*model.FiltrosPlanoAcao, error) {
	estados := make([]string, 0)
	impactos := make([]string, 0)
	problemas := make([]string, 0)
	acoes := make([]string, 0)

	r.db.Model(&model.PlanoAcao{}).Where("reservatorio_id = ?", reservatorioID).Distinct().Pluck("estado_seca", &estados)
	r.db.Model(&model.PlanoAcao{}).Where("reservatorio_id = ?", reservatorioID).Distinct().Pluck("tipos_impactos", &impactos)
	r.db.Model(&model.PlanoAcao{}).Where("reservatorio_id = ?", reservatorioID).Distinct().Pluck("problemas", &problemas)
	r.db.Model(&model.PlanoAcao{}).Where("reservatorio_id = ?", reservatorioID).Distinct().Pluck("acoes", &acoes)

	return &model.FiltrosPlanoAcao{
		Estados:   estados,
		Impactos:  impactos,
		Problemas: problemas,
		Acoes:     acoes,
	}, nil
}

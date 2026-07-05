package repository

import (
	"testing"

	"github.com/guiezz/dashboard-api/model"
	"github.com/guiezz/dashboard-api/model/simulador"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao conectar no banco de teste: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Usuario{},
		&model.HistoricoAcao{},
		&model.Reservatorio{},
		&model.Monitoramento{},
		&model.UsoAgua{},
		&model.BalancoMensal{},
		&model.ComposicaoDemanda{},
		&model.OfertaDemanda{},
		&model.PlanoAcao{},
		&model.VolumeMeta{},
		&model.Responsavel{},
		&simulador.SimAcude{},
		&simulador.SimCAV{},
		&simulador.SimEvaporacao{},
		&simulador.SimVazao{},
	); err != nil {
		t.Fatalf("falha ao migrar banco de teste: %v", err)
	}
	return db
}

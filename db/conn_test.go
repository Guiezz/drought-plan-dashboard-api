package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
)

// Sem statements preparados no servidor, uma mudança de schema (colunas novas)
// não deixa planos em cache inválidos no pooler ("cached plan must not change result type").
func TestDialector_UsaProtocoloSimples(t *testing.T) {
	d, ok := dialector("host=localhost").(*postgres.Dialector)
	assert.True(t, ok)
	assert.True(t, d.Config.PreferSimpleProtocol)
	assert.Equal(t, "host=localhost", d.Config.DSN)
}

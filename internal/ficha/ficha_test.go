package ficha

import (
	"os"
	"strings"
	"testing"

	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const header = "nome;bacia;area_drenagem_km2;ano_construcao;construtor;rio_barrado;municipios_atendidos\n"

func ptr[T any](v T) *T { return &v }

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Reservatorio{}))
	return db
}

func TestParse_Valido(t *testing.T) {
	regs, err := Parse(strings.NewReader(header +
		"Acarape do Meio;Bacias Metropolitanas;206,20;1924;IFOCS (atual DNOCS);;Redenção|Acarape\n"))
	require.NoError(t, err)
	require.Len(t, regs, 1)
	r := regs[0]
	assert.Equal(t, "Acarape do Meio", r.Nome)
	assert.Equal(t, "Bacias Metropolitanas", *r.Bacia)
	assert.Equal(t, 206.2, *r.AreaDrenagemKm2)
	assert.Equal(t, 1924, *r.AnoConstrucao)
	assert.Equal(t, "IFOCS (atual DNOCS)", *r.Construtor)
	assert.Nil(t, r.RioBarrado)
	assert.Equal(t, "Redenção; Acarape", *r.MunicipiosAtendidos)
}

func TestParse_DecimalComPontoEMilhar(t *testing.T) {
	regs, err := Parse(strings.NewReader(header + "Arneiroz II;;5.401,60;;;;\nCastro;;364.03;;;;\n"))
	require.NoError(t, err)
	assert.Equal(t, 5401.6, *regs[0].AreaDrenagemKm2)
	assert.Equal(t, 364.03, *regs[1].AreaDrenagemKm2)
}

func TestParse_CabecalhoInvalido(t *testing.T) {
	_, err := Parse(strings.NewReader("nome;outra\nX;Y\n"))
	assert.ErrorContains(t, err, "cabeçalho")
}

func TestParse_NumeroInvalido(t *testing.T) {
	_, err := Parse(strings.NewReader(header + "Acarape do Meio;;abc;;;;\n"))
	assert.ErrorContains(t, err, "linha 2")
}

func TestParse_NomeVazio(t *testing.T) {
	_, err := Parse(strings.NewReader(header + ";;;;;;\n"))
	assert.ErrorContains(t, err, "linha 2")
}

func TestAplicar_AtualizaPorNome(t *testing.T) {
	db := setupDB(t)
	require.NoError(t, db.Create(&model.Reservatorio{Nome: "A"}).Error)
	require.NoError(t, db.Create(&model.Reservatorio{Nome: "B", Construtor: ptr("DNOCS")}).Error)

	atualizados, desconhecidos, err := Aplicar(db, []Registro{{Nome: "A", AnoConstrucao: ptr(1924)}, {Nome: "Z"}})
	require.NoError(t, err)
	assert.Equal(t, 1, atualizados)
	assert.Equal(t, []string{"Z"}, desconhecidos)

	var a, b model.Reservatorio
	db.Where("nome = ?", "A").First(&a)
	db.Where("nome = ?", "B").First(&b)
	assert.Equal(t, 1924, *a.AnoConstrucao)
	assert.Equal(t, "DNOCS", *b.Construtor)
}

func TestAplicar_Idempotente(t *testing.T) {
	db := setupDB(t)
	require.NoError(t, db.Create(&model.Reservatorio{Nome: "A"}).Error)
	regs := []Registro{{Nome: "A", RioBarrado: ptr("Pacoti")}}

	for i := 0; i < 2; i++ {
		atualizados, _, err := Aplicar(db, regs)
		require.NoError(t, err)
		assert.Equal(t, 1, atualizados)
	}
	var a model.Reservatorio
	db.Where("nome = ?", "A").First(&a)
	assert.Equal(t, "Pacoti", *a.RioBarrado)
}

func TestAplicar_VazioLimpa(t *testing.T) {
	db := setupDB(t)
	require.NoError(t, db.Create(&model.Reservatorio{Nome: "A", RioBarrado: ptr("Pacoti")}).Error)

	_, _, err := Aplicar(db, []Registro{{Nome: "A"}})
	require.NoError(t, err)

	var a model.Reservatorio
	db.Where("nome = ?", "A").First(&a)
	assert.Nil(t, a.RioBarrado)
}

// Garante que o CSV versionado continua válido e cobre os 17 reservatórios.
func TestCSVVersionado(t *testing.T) {
	f, err := os.Open("../../dados_importacao/ficha_tecnica.csv")
	require.NoError(t, err)
	defer f.Close()

	regs, err := Parse(f)
	require.NoError(t, err)
	assert.Len(t, regs, 17)

	porNome := map[string]Registro{}
	for _, r := range regs {
		porNome[r.Nome] = r
	}
	assert.Equal(t, 206.2, *porNome["Acarape do Meio"].AreaDrenagemKm2)
	assert.Equal(t, 7700.0, *porNome["Fogareiro"].AreaDrenagemKm2)
	assert.Equal(t, 5401.6, *porNome["Arneiroz II"].AreaDrenagemKm2)
	for _, r := range regs {
		if r.AreaDrenagemKm2 != nil {
			assert.Greater(t, *r.AreaDrenagemKm2, 10.0, r.Nome)
		}
	}
}

func TestParse_MilharSemVirgulaEhAmbiguo(t *testing.T) {
	_, err := Parse(strings.NewReader(header + "Fogareiro;;7.700;;;;\n"))
	assert.ErrorContains(t, err, "linha 2")
}

// Dry-run não pode deixar rastro: nem colunas novas nem dados.
func TestExecutar_DryRunNaoAlteraSchemaNemDados(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE reservatorio (id integer primary key, nome text)").Error)
	require.NoError(t, db.Exec("INSERT INTO reservatorio (id, nome) VALUES (1, 'A')").Error)

	atualizados, _, err := Executar(db, []Registro{{Nome: "A", AnoConstrucao: ptr(1924)}}, true)
	require.NoError(t, err)
	assert.Equal(t, 1, atualizados)
	assert.False(t, db.Migrator().HasColumn(&model.Reservatorio{}, "ano_construcao"))
}

func TestExecutar_GravaDeVerdade(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE reservatorio (id integer primary key, nome text)").Error)
	require.NoError(t, db.Exec("INSERT INTO reservatorio (id, nome) VALUES (1, 'A')").Error)

	_, _, err = Executar(db, []Registro{{Nome: "A", AnoConstrucao: ptr(1924)}}, false)
	require.NoError(t, err)
	var a model.Reservatorio
	require.NoError(t, db.Where("nome = ?", "A").First(&a).Error)
	assert.Equal(t, 1924, *a.AnoConstrucao)
}

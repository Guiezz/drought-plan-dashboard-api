// Package ficha carrega a ficha técnica dos reservatórios a partir de um CSV
// (dados_importacao/ficha_tecnica.csv) e grava os campos no banco por nome.
package ficha

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/guiezz/dashboard-api/model"
	"gorm.io/gorm"
)

var colunas = []string{
	"nome", "bacia", "area_drenagem_km2", "ano_construcao", "construtor", "rio_barrado", "municipios_atendidos",
}

// AbrirCSV abre o CSV confinado ao diretório dele (os.OpenInRoot): um link
// simbólico ou caminho que leve para fora desse diretório é recusado.
func AbrirCSV(caminho string) (*os.File, error) {
	return os.OpenInRoot(filepath.Dir(caminho), filepath.Base(caminho))
}

// Registro é uma linha do CSV. Campos vazios viram nil.
type Registro struct {
	Nome                string
	Bacia               *string
	AreaDrenagemKm2     *float64
	AnoConstrucao       *int
	Construtor          *string
	RioBarrado          *string
	MunicipiosAtendidos *string
}

func texto(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// "7.700" sem vírgula é ambíguo (7,7 ou 7700?) e é recusado.
var milharSemVirgula = regexp.MustCompile(`^\d{1,3}(\.\d{3})+$`)

// Aceita "206,20", "5.401,60", "364.03" e "7700".
func decimal(s string) (*float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	if milharSemVirgula.MatchString(s) {
		return nil, fmt.Errorf("valor ambíguo %q: use vírgula decimal (ex.: 7.700,00) ou sem separador", s)
	}
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func inteiro(s string) (*int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// Parse lê o CSV separado por ";". Em municipios_atendidos os municípios
// vêm separados por "|" e são gravados separados por "; ".
func Parse(r io.Reader) ([]Registro, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.FieldsPerRecord = len(colunas)

	cabecalho, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("ficha: cabeçalho ilegível: %w", err)
	}
	for i, c := range colunas {
		if strings.TrimSpace(strings.TrimPrefix(cabecalho[i], "\uFEFF")) != c {
			return nil, fmt.Errorf("ficha: cabeçalho inválido, esperado %q", strings.Join(colunas, ";"))
		}
	}

	var regs []Registro
	for linha := 2; ; linha++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("ficha: linha %d: %w", linha, err)
		}
		nome := strings.TrimSpace(row[0])
		if nome == "" {
			return nil, fmt.Errorf("ficha: linha %d: nome vazio", linha)
		}
		area, err := decimal(row[2])
		if err != nil {
			return nil, fmt.Errorf("ficha: linha %d: área de drenagem inválida %q", linha, row[2])
		}
		ano, err := inteiro(row[3])
		if err != nil {
			return nil, fmt.Errorf("ficha: linha %d: ano de construção inválido %q", linha, row[3])
		}
		var municipios *string
		if m := texto(row[6]); m != nil {
			partes := strings.Split(*m, "|")
			for i := range partes {
				partes[i] = strings.TrimSpace(partes[i])
			}
			juntos := strings.Join(partes, "; ")
			municipios = &juntos
		}
		regs = append(regs, Registro{
			Nome:                nome,
			Bacia:               texto(row[1]),
			AreaDrenagemKm2:     area,
			AnoConstrucao:       ano,
			Construtor:          texto(row[4]),
			RioBarrado:          texto(row[5]),
			MunicipiosAtendidos: municipios,
		})
	}
	return regs, nil
}

// Aplicar grava a ficha por nome exato, numa transação. O CSV é a fonte da
// verdade: campo vazio grava NULL. Nomes sem reservatório voltam em desconhecidos.
func Aplicar(db *gorm.DB, regs []Registro) (atualizados int, desconhecidos []string, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, r := range regs {
			res := tx.Model(&model.Reservatorio{}).Where("nome = ?", r.Nome).Updates(map[string]any{
				"bacia":                r.Bacia,
				"area_drenagem_km2":    r.AreaDrenagemKm2,
				"ano_construcao":       r.AnoConstrucao,
				"construtor":           r.Construtor,
				"rio_barrado":          r.RioBarrado,
				"municipios_atendidos": r.MunicipiosAtendidos,
			})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				desconhecidos = append(desconhecidos, r.Nome)
				continue
			}
			atualizados++
		}
		return nil
	})
	if err != nil {
		return 0, nil, err
	}
	return atualizados, desconhecidos, nil
}

var errDryRun = errors.New("dry-run")

// Executar cria as colunas da ficha (se faltarem) e grava os registros numa
// única transação. Em dry-run tudo é desfeito no fim — inclusive as colunas,
// já que o DDL é transacional no Postgres e no SQLite.
func Executar(db *gorm.DB, regs []Registro, dryRun bool) (atualizados int, desconhecidos []string, err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.Reservatorio{}); err != nil {
			return fmt.Errorf("ficha: migrar tabela reservatorio: %w", err)
		}
		var errAplicar error
		atualizados, desconhecidos, errAplicar = Aplicar(tx, regs)
		if errAplicar != nil {
			return errAplicar
		}
		if dryRun {
			return errDryRun
		}
		return nil
	})
	if errors.Is(err, errDryRun) {
		return atualizados, desconhecidos, nil
	}
	if err != nil {
		return 0, nil, err
	}
	return atualizados, desconhecidos, nil
}

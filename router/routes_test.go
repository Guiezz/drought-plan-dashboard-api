package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/controller"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/internal/funceme"
	"github.com/guiezz/dashboard-api/middleware"
	"github.com/guiezz/dashboard-api/model"
	"github.com/guiezz/dashboard-api/model/simulador"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupE2E(t *testing.T) (*gin.Engine, *gorm.DB, []byte) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
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

	reservatorioRepo := repository.NewReservatorioRepository(db)
	planoAcaoRepo := repository.NewPlanoAcaoRepository(db)
	balancoRepo := repository.NewBalancoHidricoRepository(db)
	usoRepo := repository.NewUsoAguaRepository(db)
	respRepo := repository.NewResponsavelRepository(db)
	simulacaoRepo := repository.NewSimulacaoRepository(db)

	secaCalc := calculator.NewSecaCalculator()
	funcemeSvc := funceme.NewFuncemeService("http://test.com")
	simulacaoCalc := calculator.NewSimuladorHidrico()

	reservatorioUseCase := usecase.NewReservatorioUseCase(reservatorioRepo, planoAcaoRepo, secaCalc, funcemeSvc)
	planoAcaoUseCase := usecase.NewPlanoAcaoUseCase(planoAcaoRepo)
	balancoUseCase := usecase.NewBalancoHidricoUseCase(balancoRepo)
	usoUseCase := usecase.NewUsoAguaUseCase(usoRepo)
	respUseCase := usecase.NewResponsavelUseCase(respRepo)
	simulacaoUseCase := usecase.NewSimulacaoUseCase(simulacaoRepo, simulacaoCalc)

	reservatorioController := controller.NewReservatorioController(reservatorioUseCase)
	planoAcaoController := controller.NewPlanoAcaoController(planoAcaoUseCase)
	balancoController := controller.NewBalancoHidricoController(balancoUseCase)
	usoController := controller.NewUsoAguaController(usoUseCase)
	respController := controller.NewResponsavelController(respUseCase)
	simulacaoController := controller.NewSimulacaoController(simulacaoUseCase)

	secret := []byte("test-secret")
	authController := controller.NewAuthController(db, secret)

	loginLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
	simulacaoLimiter := middleware.NewRateLimiter(100, 1*time.Minute)

	frontendURLs := []string{"http://localhost:3000"}
	corsCtrl := controller.NewCorsController(frontendURLs)

	router := SetupRouter(
		reservatorioController,
		planoAcaoController,
		balancoController,
		usoController,
		respController,
		simulacaoController,
		authController,
		corsCtrl,
		frontendURLs,
		loginLimiter,
		simulacaoLimiter,
	)

	return router, db, secret
}

func TestE2E_Ping(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "pong!", resp["message"])
}

func TestE2E_ReservatoriosListaVazia(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reservatorios", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res []model.Reservatorio
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	assert.Empty(t, res)
}

func TestE2E_ReservatoriosComDados(t *testing.T) {
	router, db, _ := setupE2E(t)

	db.Create(&model.Reservatorio{Nome: "Teste", Capacidadehm3: 100})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reservatorios", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res []model.Reservatorio
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	require.Len(t, res, 1)
	assert.Equal(t, "Teste", res[0].Nome)
}

func TestE2E_LoginSucesso(t *testing.T) {
	router, db, _ := setupE2E(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("senha123"), bcrypt.DefaultCost)
	db.Create(&model.Usuario{Nome: "Admin", Email: "admin@test.com", SenhaHash: string(hash), Role: "cogerh"})

	w := httptest.NewRecorder()
	body := `{"email":"admin@test.com","senha":"senha123"}`
	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["token"])
	assert.Equal(t, "Admin", resp["usuario"].(map[string]interface{})["nome"])
}

func TestE2E_LoginFalha(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	body := `{"email":"naoexiste@test.com","senha":"senha"}`
	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestE2E_RotaProtegidaSemToken(t *testing.T) {
	router, db, _ := setupE2E(t)

	db.Create(&model.Reservatorio{Nome: "Teste", Capacidadehm3: 100})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/reservatorios/1/action-plans/1/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestE2E_RotaProtegidaComToken(t *testing.T) {
	router, db, _ := setupE2E(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("senha123"), bcrypt.DefaultCost)
	db.Create(&model.Usuario{Nome: "Admin", Email: "admin@test.com", SenhaHash: string(hash), Role: "cogerh"})
	db.Create(&model.Reservatorio{Nome: "Teste", Capacidadehm3: 100})
	db.Create(&model.PlanoAcao{ReservatorioID: 1, Situacao: "nao_iniciado"})

	w := httptest.NewRecorder()
	body := `{"email":"admin@test.com","senha":"senha123"}`
	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var loginResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	tokenStr := loginResp["token"].(string)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/reservatorios/1/action-plans/1/status", strings.NewReader(`{"situacao":"em_andamento"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "rota protegida com token válido deve retornar 200")
}

func TestE2E_RotaNaoEncontrada(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/rota-inexistente", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestE2E_CORSPermitida(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/reservatorios", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestE2E_CORSNaoPermitida(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/api/reservatorios", nil)
	req.Header.Set("Origin", "http://evil.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	router.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestE2E_RotaPublicaFuncionaSemToken(t *testing.T) {
	router, db, _ := setupE2E(t)

	db.Create(&model.Reservatorio{Nome: "Teste", Capacidadehm3: 100})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reservatorios", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestE2E_SimulacaoAcudes(t *testing.T) {
	router, db, _ := setupE2E(t)

	db.Create(&simulador.SimAcude{Codigo: 1, Nome: "Açude Teste", Capacidade: 1000000, Municipio: "Teste", EstacaoEvapID: 1})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/simulacao/acudes", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	require.Len(t, resp, 1)
	assert.Equal(t, float64(1), resp[0]["codigo"])
}

func TestE2E_SimulacaoAnos(t *testing.T) {
	router, db, _ := setupE2E(t)

	db.Create(&simulador.SimAcude{Codigo: 1, Nome: "Açude Teste", Capacidade: 1000000, Municipio: "Teste", EstacaoEvapID: 1})
	db.Create(&simulador.SimVazao{AcudeID: 1, Ano: 2020})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/simulacao/anos?reservatorio_id=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp["anos"])
}

func TestE2E_SimulacaoRunEndpointExiste(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/simulacao/run", strings.NewReader(`{"reservatorio_id":0}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusNotFound, w.Code, "rota /api/simulacao/run deve existir")
	assert.NotEqual(t, http.StatusUnauthorized, w.Code, "rota deve ser pública")
}

func TestE2E_RotaReservatorioInexistente(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reservatorios/999/dashboard/summary", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestE2E_LoginJSONInvalido(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", strings.NewReader(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestE2E_CorsCheck_Livre(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cors-check", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["allowed_origins"], "http://localhost:3000")
	assert.Equal(t, "", resp["request_origin"])
}

func TestE2E_CorsCheck_ComOrigin(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/cors-check", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Host", "test.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["allowed_origins"], "http://localhost:3000")
	assert.Equal(t, "http://localhost:3000", resp["request_origin"])
}

func TestE2E_CorsCheck_Api(t *testing.T) {
	router, _, _ := setupE2E(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/cors-check", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Contains(t, resp["allowed_origins"], "http://localhost:3000")
}

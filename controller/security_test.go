package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/internal/calculator"
	"github.com/guiezz/dashboard-api/middleware"
	"github.com/guiezz/dashboard-api/model"
	"github.com/guiezz/dashboard-api/repository"
	"github.com/guiezz/dashboard-api/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSecurityTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	if err := db.AutoMigrate(&model.Usuario{}, &model.PlanoAcao{}, &model.HistoricoAcao{}, &model.Reservatorio{}); err != nil {
		t.Fatalf("falha ao migrar banco de teste: %v", err)
	}
	return db
}

func TestSecurity_ProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")

	t.Run("rota protegida sem token retorna 401", func(t *testing.T) {
		db := setupSecurityTestDB(t)
		_ = NewAuthController(db, secret)
		router := gin.New()

		protegido := router.Group("/")
		protegido.Use(middleware.AuthMiddleware(secret))
		protegido.PUT("/action-plans/:acaoId/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/action-plans/1/status", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("rota protegida com token valido retorna 200", func(t *testing.T) {
		db := setupSecurityTestDB(t)
		ac := NewAuthController(db, secret)

		hash, _ := bcrypt.GenerateFromPassword([]byte("senha"), bcrypt.DefaultCost)
		db.Create(&model.Usuario{Nome: "Admin", Email: "admin@test.com", SenhaHash: string(hash), Role: "cogerh"})

		router := gin.New()
		router.POST("/login", ac.Login)

		protegido := router.Group("/")
		protegido.Use(middleware.AuthMiddleware(secret))
		protegido.GET("/me", func(c *gin.Context) {
			id, _ := c.Get("usuario_id")
			role, _ := c.Get("role")
			c.JSON(http.StatusOK, gin.H{"usuario_id": id, "role": role})
		})

		w := httptest.NewRecorder()
		body := `{"email":"admin@test.com","senha":"senha"}`
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)

		var loginResp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
		tokenStr := loginResp["token"].(string)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var meResp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &meResp)
		assert.NotEmpty(t, meResp["usuario_id"])
	})
}

func TestSecurity_ErroNaoVazaDetalhes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("erro 500 retorna mensagem genérica", func(t *testing.T) {
		router := gin.New()
		router.GET("/forcar-erro", func(c *gin.Context) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar reservatórios"})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/forcar-erro", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		msg := resp["error"].(string)
		assert.NotContains(t, msg, "record not found")
		assert.NotContains(t, msg, "nil pointer")
		assert.NotContains(t, msg, "runtime")
	})
}

func TestSecurity_CORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("origem permitida recebe headers CORS", func(t *testing.T) {
		router := gin.New()
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://meuapp.com"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Origin"},
			AllowCredentials: true,
		}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://meuapp.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		router.ServeHTTP(w, req)

		assert.Equal(t, "http://meuapp.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("origem nao permitida nao recebe headers CORS", func(t *testing.T) {
		router := gin.New()
		router.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://meuapp.com"},
			AllowMethods:     []string{"GET"},
			AllowHeaders:     []string{"Origin"},
			AllowCredentials: true,
		}))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://evil.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		router.ServeHTTP(w, req)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestSecurity_RotaPublicaFuncionaSemToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupSecurityTestDB(t)
	db.Create(&model.Reservatorio{Nome: "Público", Capacidadehm3: 100})

	repo := repository.NewReservatorioRepository(db)
	planoRepo := repository.NewPlanoAcaoRepository(db)
	secaCalc := calculator.NewSecaCalculator()
	reservatorioUseCase := usecase.NewReservatorioUseCase(repo, planoRepo, secaCalc, nil)
	resCtrl := NewReservatorioController(reservatorioUseCase)

	router := gin.New()
	router.GET("/api/reservatorios", resCtrl.GetReservatorios)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reservatorios", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var res []model.Reservatorio
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	assert.Len(t, res, 1)
}

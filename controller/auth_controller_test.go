package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guiezz/dashboard-api/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao conectar no banco de teste: %v", err)
	}
	db.AutoMigrate(&model.Usuario{})
	return db
}

func seedUser(db *gorm.DB, email, password, role string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	db.Create(&model.Usuario{
		Nome:      "Teste",
		Email:     email,
		SenhaHash: string(hash),
		Role:      role,
	})
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("credenciais corretas retorna token", func(t *testing.T) {
		db := setupTestDB(t)
		seedUser(db, "teste@test.com", "senha123", "cogerh")

		ac := NewAuthController(db, []byte("test-secret"))
		router := gin.New()
		router.POST("/login", ac.Login)

		body := `{"email":"teste@test.com","senha":"senha123"}`
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotEmpty(t, resp["token"])
		assert.NotNil(t, resp["usuario"])
	})

	t.Run("email não encontrado retorna 401", func(t *testing.T) {
		db := setupTestDB(t)
		ac := NewAuthController(db, []byte("test-secret"))
		router := gin.New()
		router.POST("/login", ac.Login)

		body := `{"email":"naoexiste@test.com","senha":"senha123"}`
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("senha errada retorna 401", func(t *testing.T) {
		db := setupTestDB(t)
		seedUser(db, "teste@test.com", "senha123", "cogerh")

		ac := NewAuthController(db, []byte("test-secret"))
		router := gin.New()
		router.POST("/login", ac.Login)

		body := `{"email":"teste@test.com","senha":"senha_errada"}`
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("JSON inválido retorna 400", func(t *testing.T) {
		db := setupTestDB(t)
		ac := NewAuthController(db, []byte("test-secret"))
		router := gin.New()
		router.POST("/login", ac.Login)

		body := `{"email":"teste@test.com"}`
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

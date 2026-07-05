package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	secret := []byte("test-secret")

	t.Run("token ausente retorna 401", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("formato invalido retorna 401", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Token invalido")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token invalido retorna 401", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Bearer token.qualquer.invalido")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token valido passa e define claims no contexto", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), func(c *gin.Context) {
			id, _ := c.Get("usuario_id")
			role, _ := c.Get("role")
			assert.Equal(t, uint(42), id)
			assert.Equal(t, "admin_cogerh", role)
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"usuario_id": float64(42),
			"role":       "admin_cogerh",
			"exp":        time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(secret)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("token expirado retorna 401", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"usuario_id": float64(1),
			"role":       "cogerh",
			"exp":        time.Now().Add(-1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(secret)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("algoritmo diferente rejeitado", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"usuario_id": float64(1),
			"role":       "cogerh",
			"exp":        time.Now().Add(1 * time.Hour).Unix(),
		})
		token.Header["alg"] = "RS256"
		tokenString, _ := token.SignedString(secret)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token com assinatura de segredo diferente retorna 401", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.GET("/protegido", AuthMiddleware(secret), handlerOK)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"usuario_id": float64(1),
			"role":       "cogerh",
			"exp":        time.Now().Add(1 * time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("outro-secret"))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/protegido", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func handlerOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func setupRateLimiterTest(limit int, window time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(limit, window)
	r := gin.New()
	r.POST("/test", rl.Middleware(), handlerOK)
	return r
}

func TestRateLimiter(t *testing.T) {
	t.Run("permite ate o limite e bloqueia o excedente", func(t *testing.T) {
		router := setupRateLimiterTest(3, 1*time.Minute)

		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", nil)
			req.RemoteAddr = "192.168.1.1:8080"
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "requisição %d deveria passar", i+1)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "192.168.1.1:8080"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("ips diferentes tem contadores independentes", func(t *testing.T) {
		router := setupRateLimiterTest(2, 1*time.Minute)

		ips := []string{"10.0.0.1:1234", "10.0.0.2:1234", "10.0.0.3:1234"}
		for _, ip := range ips {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/test", nil)
			req.RemoteAddr = ip
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code, "IP %s deveria passar", ip)
		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "10.0.0.1 tem apenas 1 req, deveria passar")
	})

	t.Run("reseta apos janela de tempo", func(t *testing.T) {
		router := setupRateLimiterTest(1, 50*time.Millisecond)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)

		time.Sleep(60 * time.Millisecond)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "após reset, deveria passar novamente")
	})
}

func TestRateLimiter_RotaSemLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/livre", handlerOK)

	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/livre", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

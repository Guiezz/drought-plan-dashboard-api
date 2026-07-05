package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CorsController struct {
	FrontendURLs []string
}

func NewCorsController(frontendURLs []string) *CorsController {
	return &CorsController{FrontendURLs: frontendURLs}
}

func (ctrl *CorsController) CheckCORS(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"allowed_origins": ctrl.FrontendURLs,
		"request_origin":  c.GetHeader("Origin"),
		"request_host":    c.GetHeader("Host"),
	})
}

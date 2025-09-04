package controllers

import (
	"common/domain/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthController estructura para manejar la ruta de Health
type HealthController struct {
}

// NewHealthController constructor para HealthController
func NewHealthController() *HealthController {
	return &HealthController{}
}

// GetHealth
func (c *HealthController) GetHealth(ctx *gin.Context) {

	entry := logger.FromContext(ctx)

	entry.Info("Health check")

	// Responder con un JSON que contiene la URL generada
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "ok",
			"message":   "The service is online and functioning properly.",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

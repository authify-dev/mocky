package health

import (
	"mocky/internal/api/health/interface/controllers"

	"github.com/gin-gonic/gin"
)

func SetupHealthModule(r *gin.Engine) {

	healthController := controllers.NewHealthController()

	// Rutas de health
	health := r.Group("/v1/health")

	health.GET("", healthController.GetHealth)
}

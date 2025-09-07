package health

import (
	"mocky/internal/api/health/interface/controllers"
	"mocky/internal/core/settings"

	"github.com/gin-gonic/gin"
)

func SetupHealthModule(r *gin.Engine) {

	healthController := controllers.NewHealthController()

	// Rutas de health
	health := r.Group(settings.Settings.ROOT_PATH + "/health")

	health.GET("", healthController.GetHealth)
}

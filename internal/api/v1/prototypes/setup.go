package prototypes

import (
	"mocky/internal/api/v1/prototypes/app/services"
	"mocky/internal/api/v1/prototypes/interface/controllers"
	validator_controller "mocky/internal/context/controllers"
	"mocky/internal/context/controllers/placeholder"
	"mocky/internal/core/settings"
	prototypes "mocky/internal/db/mongo/prototypes"

	"github.com/gin-gonic/gin"
)

func SetupPrototypesModule(r *gin.Engine) {

	// repositories
	prototypesRepository := prototypes.NewPrototypesMongoRepository(
		settings.Settings.MONGO_DSN,
		"mocky_db",
		"prototypes",
	)

	// Validator
	validator := validator_controller.NewValidator()

	// Placeholder
	placeholderController := placeholder.NewPlaceholderController()

	// Services
	prototypesService := services.NewPrototypesService(prototypesRepository, validator, placeholderController)

	// Controllers
	prototypesController := controllers.NewPrototypesController(prototypesService)

	// Routes
	prototypesGroup := r.Group("/v1/prototypes")
	prototypesGroup.POST("", prototypesController.Create)
	prototypesGroup.GET("", prototypesController.List)
	prototypesGroup.GET("/:id", prototypesController.Retrieve)

	mockyGroup := r.Group("/v1/mocky")
	mockyGroup.Any("/*path", prototypesController.Mock)

}

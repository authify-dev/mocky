package prototypes

import (
	"mocky/internal/api/v1/prototypes/app/services"
	"mocky/internal/api/v1/prototypes/interface/controllers"
	validator_controller "mocky/internal/context/controllers"
	"mocky/internal/context/controllers/placeholder"
	"mocky/internal/core/settings"
	prototypes_inmemory "mocky/internal/db/inmemory/prototypes"
	prototypes "mocky/internal/db/mongo/prototypes"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupPrototypesModule(r *gin.Engine) {

	// repositories
	// prototypesRepository := prototypes.NewPrototypesMongoRepository(
	// 	settings.Settings.MONGO_DSN,
	// 	"mocky_db",
	// 	"prototypes",
	// )

	prototypesRepositoryInMemory := prototypes_inmemory.NewInMemoryPrototypesRepository(
		func(m prototypes.PrototypeModel) prototypes.PrototypeListModel {
			return prototypes.PrototypeListModel{
				ID:        m.ID,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
				Request: prototypes.RequestListView{
					Method:  m.Request.Method,
					UrlPath: m.Request.UrlPath,
				},
				Name: m.Name,
				// … completa según tu struct
			}
		},
		15*time.Minute,
	)

	// Validator
	validator := validator_controller.NewValidator()

	// Placeholder
	placeholderController := placeholder.NewPlaceholderController()

	// Services
	prototypesService := services.NewPrototypesService(prototypesRepositoryInMemory, validator, placeholderController)

	// Controllers
	prototypesController := controllers.NewPrototypesController(prototypesService)

	// Routes
	prototypesGroup := r.Group(settings.Settings.ROOT_PATH + "/v1/prototypes")
	prototypesGroup.POST("", prototypesController.Create)
	prototypesGroup.GET("", prototypesController.List)
	prototypesGroup.GET("/:id", prototypesController.Retrieve)

	mockyGroup := r.Group(settings.Settings.ROOT_PATH + "/v1/mocky")
	mockyGroup.Any("/*path", prototypesController.Mock)

}

package services

import (
	validator_controller "mocky/internal/context/controllers"
	"mocky/internal/context/controllers/placeholder"
	prototypes "mocky/internal/db/mongo/prototypes"
)

type PrototypesService struct {
	prototypesRepository  *prototypes.PrototypesMongoRepository
	validator             *validator_controller.ValidatorRequest
	placeholderController *placeholder.PlaceholderController
}

func NewPrototypesService(
	prototypesRepository *prototypes.PrototypesMongoRepository,
	validator *validator_controller.ValidatorRequest,
	placeholderController *placeholder.PlaceholderController,
) *PrototypesService {
	return &PrototypesService{
		prototypesRepository:  prototypesRepository,
		validator:             validator,
		placeholderController: placeholderController,
	}
}

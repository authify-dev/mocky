package services

import (
	"mocky/internal/api/v1/prototypes/domain/repositories"
	validator_controller "mocky/internal/context/controllers"
	"mocky/internal/context/controllers/placeholder"
)

type PrototypesService struct {
	prototypesRepository  repositories.RepositoryPrototypes
	validator             *validator_controller.ValidatorRequest
	placeholderController *placeholder.PlaceholderController
}

func NewPrototypesService(
	prototypesRepository repositories.RepositoryPrototypes,
	validator *validator_controller.ValidatorRequest,
	placeholderController *placeholder.PlaceholderController,
) *PrototypesService {
	return &PrototypesService{
		prototypesRepository:  prototypesRepository,
		validator:             validator,
		placeholderController: placeholderController,
	}
}

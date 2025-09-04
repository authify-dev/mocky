package services

import (
	validator_controller "mocky/internal/context/controllers"
	prototypes "mocky/internal/db/mongo/prototypes"
)

type PrototypesService struct {
	prototypesRepository *prototypes.PrototypesMongoRepository
	validator            *validator_controller.ValidatorRequest
}

func NewPrototypesService(prototypesRepository *prototypes.PrototypesMongoRepository, validator *validator_controller.ValidatorRequest) *PrototypesService {
	return &PrototypesService{
		prototypesRepository: prototypesRepository,
		validator:            validator,
	}
}

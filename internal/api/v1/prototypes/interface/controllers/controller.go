package controllers

import "mocky/internal/api/v1/prototypes/app/services"

type PrototypesController struct {
	prototypesService *services.PrototypesService
}

func NewPrototypesController(prototypesService *services.PrototypesService) *PrototypesController {
	return &PrototypesController{
		prototypesService: prototypesService,
	}
}

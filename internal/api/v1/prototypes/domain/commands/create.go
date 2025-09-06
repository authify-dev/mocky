package commands

import "mocky/internal/api/v1/prototypes/domain/entities"

type CreatePrototypeCommand struct {
	Request  entities.RequestEntity  `json:"request" binding:"required"`
	Response entities.ResponseEntity `json:"response" binding:"required"`
	Name     string                  `json:"name" binding:"required"`
}

func (c CreatePrototypeCommand) Validate() error {
	return nil
}

func (c CreatePrototypeCommand) ToEntity() entities.PrototypeEntity {
	return entities.PrototypeEntity{
		Name:     c.Name,
		Request:  c.Request,
		Response: c.Response,
	}
}

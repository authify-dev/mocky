package dtos

import (
	"common/utils/ctypes"
	"errors"
	"mocky/internal/api/v1/prototypes/domain/commands"
	"mocky/internal/api/v1/prototypes/domain/entities"
)

type CreatePrototypeDTO struct {
	Request RequestDTO `json:"request" binding:"required"`
}

func (dto CreatePrototypeDTO) Validate() error {

	if dto.Request.Validate() != nil {
		return errors.New("request is invalid: " + dto.Request.Validate().Error())
	}

	return nil
}

type RequestDTO struct {
	Method     string            `json:"method" binding:"required"`
	UrlPath    string            `json:"urlPath" binding:"required"`
	Headers    map[string]string `json:"headers"`
	BodySchema BodySchemaDTO     `json:"bodySchema"`
}

func (dto RequestDTO) Validate() error {

	if dto.Method == "" {
		return errors.New("method is required")
	}

	if dto.UrlPath == "" {
		return errors.New("urlPath is required")
	}

	if dto.BodySchema.Validate() != nil {
		return errors.New("bodySchema is invalid: " + dto.BodySchema.Validate().Error())
	}

	return nil
}

func (dto RequestDTO) ToEntity() entities.RequestEntity {

	return entities.RequestEntity{
		Method:     dto.Method,
		UrlPath:    dto.UrlPath,
		Headers:    dto.Headers,
		BodySchema: dto.BodySchema.ToEntity(),
	}
}

type BodySchemaDTO struct {
	Name                string        `json:"name" binding:"required"`
	TypeSchema          string        `json:"type_schema" binding:"required"`
	AditionalProperties bool          `json:"aditional_properties"`
	Properties          []PropertyDTO `json:"properties"`
}

func (dto BodySchemaDTO) Validate() error {

	if dto.Name == "" {
		return errors.New("name is required")
	}

	if dto.TypeSchema == "" {
		return errors.New("typeSchema is required")
	}

	for _, property := range dto.Properties {
		if property.Validate() != nil {
			return errors.New("properties is invalid: " + property.Validate().Error())
		}
	}

	return nil
}

func (dto BodySchemaDTO) ToEntity() entities.BodySchemaEntity {

	return entities.BodySchemaEntity{
		Name:                dto.Name,
		TypeSchema:          dto.TypeSchema,
		AditionalProperties: dto.AditionalProperties,
		Properties: ctypes.Map(dto.Properties, func(property PropertyDTO) entities.PropertyEntity {
			return property.ToEntity()
		}),
	}
}

type PropertyDTO struct {
	Name       string        `json:"name" binding:"required"`
	IsRequired bool          `json:"is_required"`
	Type       string        `json:"type" binding:"required"`
	MinLength  int32         `json:"min_length"`
	MaxLength  int32         `json:"max_length"`
	Format     string        `json:"format"`
	Pattern    string        `json:"pattern"`
	Properties []PropertyDTO `json:"properties"`
}

func (dto PropertyDTO) Validate() error {

	if dto.Name == "" {
		return errors.New("name is required")
	}

	if dto.Type == "" {
		return errors.New("type is required")
	}

	return nil
}

func (dto PropertyDTO) ToEntity() entities.PropertyEntity {

	return entities.PropertyEntity{
		Name:       dto.Name,
		IsRequired: dto.IsRequired,
		Type:       dto.Type,
		MinLength:  dto.MinLength,
		MaxLength:  dto.MaxLength,
		Format:     dto.Format,
		Pattern:    dto.Pattern,
		Properties: ctypes.Map(dto.Properties, func(property PropertyDTO) entities.PropertyEntity {
			return property.ToEntity()
		}),
	}
}

func (dto CreatePrototypeDTO) ToCommand() commands.CreatePrototypeCommand {
	return commands.CreatePrototypeCommand{
		Request: dto.Request.ToEntity(),
	}
}

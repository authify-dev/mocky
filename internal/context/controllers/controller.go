package validator_controller

import "mocky/internal/api/v1/prototypes/domain/entities"

type ValidatorRequest struct {
	registry map[string]ValidationFunc
}

func NewValidator() *ValidatorRequest {
	return &ValidatorRequest{
		registry: buildValidatorRegistry(),
	}
}

func (v *ValidatorRequest) Validate(schema entities.BodySchemaEntity, payload map[string]any) []ValidationError {

	errs := ValidateAgainstSchema(schema, payload, v.registry)
	if len(errs) > 0 {
		return errs
	}

	return nil
}

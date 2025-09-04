package validator_controller

import "mocky/internal/api/v1/prototypes/domain/entities"

type frame struct {
	path                 string
	schemaName           string
	additionalProperties bool
	props                []entities.PropertyEntity
	value                map[string]any
}

// Each validator returns if passed (ok) and, for "object", a child frame to push.
type ValidationFunc func(
	prop entities.PropertyEntity,
	val any,
	path string,
	parentAdditional bool,
	errs *[]ValidationError,
) (ok bool, next *frame)

package validator_controller

import (
	"fmt"
	"strings"

	"mocky/internal/api/v1/prototypes/domain/entities"
)

func joinPath(base, next string) string {
	if base == "" {
		return next
	}
	return base + "." + next
}

// ======== Validation Engine ========

func ValidateAgainstSchema(schema entities.BodySchemaEntity, payload map[string]any, registry map[string]ValidationFunc) []ValidationError {
	if registry == nil {
		// Default registry
		registry = buildValidatorRegistry()
	}

	var errs []ValidationError
	if strings.ToLower(schema.TypeSchema) != "object" {
		errs = append(errs, ValidationError{Err: "root type_schema must be 'object'"})
		return errs
	}

	propIndex := func(props []entities.PropertyEntity) map[string]entities.PropertyEntity {
		m := make(map[string]entities.PropertyEntity, len(props))
		for _, p := range props {
			m[p.Name] = p
		}
		return m
	}

	stack := []frame{{
		path:                 "",
		schemaName:           schema.Name,
		additionalProperties: schema.AditionalProperties,
		props:                schema.Properties,
		value:                payload,
	}}

	for len(stack) > 0 {
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		propsByName := propIndex(f.props)

		// Required fields
		for _, p := range f.props {
			if p.IsRequired {
				if _, ok := f.value[p.Name]; !ok {
					errs = append(errs, ValidationError{Path: joinPath(f.path, p.Name), Err: "missing required field"})
				}
			}
		}

		// Additional properties
		if !f.additionalProperties {
			for k := range f.value {
				if _, ok := propsByName[k]; !ok {
					errs = append(errs, ValidationError{Path: joinPath(f.path, k), Err: "property not allowed (additional_properties=false)"})
				}
			}
		}

		// Validate each property
		for _, p := range f.props {
			raw, ok := f.value[p.Name]
			if !ok {
				continue
			}
			propPath := joinPath(f.path, p.Name)

			validator, exists := registry[strings.ToLower(p.Type)]
			if !exists {
				errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("unsupported type in schema: %s", p.Type)})
				continue
			}

			if ok, next := validator(p, raw, propPath, f.additionalProperties, &errs); ok && next != nil {
				stack = append(stack, *next)
			}
		}
	}
	return errs
}

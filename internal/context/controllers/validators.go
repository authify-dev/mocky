package validator_controller

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"mocky/internal/api/v1/prototypes/domain/entities"
)

// ======== Validator Registry ========

func buildValidatorRegistry() map[string]ValidationFunc {
	return map[string]ValidationFunc{
		"string":  validateString,
		"boolean": validateBoolean,
		"number":  validateNumber,
		"integer": validateInteger,
		"object":  validateObject,
	}
}

// ======== Validators ========

func validateString(p entities.PropertyEntity, val any, path string, _ bool, errs *[]ValidationError) (bool, *frame) {
	s, ok := val.(string)
	if !ok {
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("expected string, got %T", val)})
		return false, nil
	}
	if p.MinLength > 0 && int32(len(s)) < p.MinLength {
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("minimum length is %d", p.MinLength)})
	}
	if p.MaxLength > 0 && int32(len(s)) > p.MaxLength {
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("maximum length is %d", p.MaxLength)})
	}
	if p.Pattern != "" {
		re, compErr := regexp.Compile(p.Pattern)
		if compErr != nil {
			*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("invalid regex in schema: %v", compErr)})
		} else if !re.MatchString(s) {
			*errs = append(*errs, ValidationError{Path: path, Err: "value does not match required pattern"})
		}
	}
	if p.Format != "" {
		switch strings.ToLower(p.Format) {
		case "email":
			if !isEmail(s) {
				*errs = append(*errs, ValidationError{Path: path, Err: "invalid email format"})
			}
		case "date":
			if !isYYYYMMDD(s) {
				*errs = append(*errs, ValidationError{Path: path, Err: "invalid date format (expected YYYY-MM-DD)"})
			}
		}
	}
	return true, nil
}

func validateBoolean(_ entities.PropertyEntity, val any, path string, _ bool, errs *[]ValidationError) (bool, *frame) {
	if _, ok := val.(bool); !ok {
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("expected boolean, got %T", val)})
		return false, nil
	}
	return true, nil
}

func validateNumber(_ entities.PropertyEntity, val any, path string, _ bool, errs *[]ValidationError) (bool, *frame) {
	switch val.(type) {
	case float64, float32, int, int32, int64:
		return true, nil
	default:
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("expected number, got %T", val)})
		return false, nil
	}
}

func validateInteger(_ entities.PropertyEntity, val any, path string, _ bool, errs *[]ValidationError) (bool, *frame) {
	switch v := val.(type) {
	case float64:
		if v != float64(int64(v)) {
			*errs = append(*errs, ValidationError{Path: path, Err: "expected integer, got float"})
			return false, nil
		}
		return true, nil
	case int, int32, int64:
		return true, nil
	default:
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("expected integer, got %T", val)})
		return false, nil
	}
}

func validateObject(p entities.PropertyEntity, val any, path string, parentAdditional bool, errs *[]ValidationError) (bool, *frame) {
	m, ok := val.(map[string]any)
	if !ok {
		*errs = append(*errs, ValidationError{Path: path, Err: fmt.Sprintf("expected object, got %T", val)})
		return false, nil
	}
	childFrame := &frame{
		path:                 path,
		schemaName:           p.Name,
		additionalProperties: parentAdditional,
		props:                p.Properties,
		value:                m,
	}
	return true, childFrame
}

// ======== Utils ========

var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func isEmail(s string) bool { return emailRe.MatchString(s) }

func isYYYYMMDD(s string) bool {
	if len(s) != 10 {
		return false
	}
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type BodySchemaEntity struct {
	Name                string           `json:"name" binding:"required"`
	TypeSchema          string           `json:"type_schema" binding:"required"` // "object"
	AditionalProperties bool             `json:"aditional_properties"`
	Properties          []PropertyEntity `json:"properties"`
}

type PropertyEntity struct {
	Name       string           `json:"name" binding:"required"`
	IsRequired bool             `json:"is_required"`
	Type       string           `json:"type" binding:"required"` // string, boolean, number, integer, object
	MinLength  int32            `json:"min_length"`
	MaxLength  int32            `json:"max_length"`
	Format     string           `json:"format"`  // email, date (YYYY-MM-DD)
	Pattern    string           `json:"pattern"` // regex
	Properties []PropertyEntity `json:"properties"`
}

// ---------------------------
// Validador con pila
// ---------------------------

type frame struct {
	path                 string
	schemaName           string
	additionalProperties bool
	props                []PropertyEntity
	value                map[string]any
}

type ValidationError struct {
	Path string
	Err  string
}

func (e ValidationError) String() string {
	if e.Path == "" {
		return e.Err
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Err)
}

func ValidateAgainstSchema(schema BodySchemaEntity, payload map[string]any) []ValidationError {
	var errs []ValidationError

	// Validación básica del esquema raíz
	if strings.ToLower(schema.TypeSchema) != "object" {
		errs = append(errs, ValidationError{Path: "", Err: "type_schema del root debe ser 'object'"})
		return errs
	}

	// Índice de propiedades por nombre para búsquedas rápidas
	propIndex := func(props []PropertyEntity) map[string]PropertyEntity {
		m := make(map[string]PropertyEntity, len(props))
		for _, p := range props {
			m[p.Name] = p
		}
		return m
	}

	stack := []frame{
		{
			path:                 "",
			schemaName:           schema.Name,
			additionalProperties: schema.AditionalProperties,
			props:                schema.Properties,
			value:                payload,
		},
	}

	for len(stack) > 0 {
		// pop
		f := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		propsByName := propIndex(f.props)

		// 1) Requeridos
		for _, p := range f.props {
			if p.IsRequired {
				if _, ok := f.value[p.Name]; !ok {
					errs = append(errs, ValidationError{
						Path: joinPath(f.path, p.Name),
						Err:  "campo requerido ausente",
					})
				}
			}
		}

		// 2) additionalProperties
		if !f.additionalProperties {
			for k := range f.value {
				if _, ok := propsByName[k]; !ok {
					errs = append(errs, ValidationError{
						Path: joinPath(f.path, k),
						Err:  "propiedad no permitida (additional_properties=false)",
					})
				}
			}
		}

		// 3) Validación por propiedad
		for _, p := range f.props {
			raw, ok := f.value[p.Name]
			if !ok {
				// ya reportamos si era requerido; si no, sigue
				continue
			}
			propPath := joinPath(f.path, p.Name)

			switch strings.ToLower(p.Type) {
			case "string":
				s, ok := raw.(string)
				if !ok {
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo esperado string, recibido %T", raw)})
					continue
				}
				// min/max length
				if p.MinLength > 0 && int32(len(s)) < p.MinLength {
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("longitud mínima %d", p.MinLength)})
				}
				// Si MaxLength==0 interpretamos como sin límite (compat con tu modelo)
				if p.MaxLength > 0 && int32(len(s)) > p.MaxLength {
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("longitud máxima %d", p.MaxLength)})
				}
				// pattern
				if p.Pattern != "" {
					re, compErr := regexp.Compile(p.Pattern)
					if compErr != nil {
						errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("pattern inválido en schema: %v", compErr)})
					} else if !re.MatchString(s) {
						errs = append(errs, ValidationError{Path: propPath, Err: "no cumple el patrón requerido"})
					}
				}
				// format
				if p.Format != "" {
					switch strings.ToLower(p.Format) {
					case "email":
						if !isEmail(s) {
							errs = append(errs, ValidationError{Path: propPath, Err: "formato email inválido"})
						}
					case "date":
						if !isYYYYMMDD(s) {
							errs = append(errs, ValidationError{Path: propPath, Err: "formato date inválido (YYYY-MM-DD)"})
						}
					default:
						// formatos que no implementamos: ignorar
					}
				}

			case "boolean":
				if _, ok := raw.(bool); !ok {
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo esperado boolean, recibido %T", raw)})
				}

			case "number":
				// Aceptamos float64 (tipo por defecto de JSON) o int
				switch raw.(type) {
				case float64, float32, int, int32, int64:
					// ok
				default:
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo esperado number, recibido %T", raw)})
				}

			case "integer":
				switch v := raw.(type) {
				case float64:
					// JSON numérico llega como float64: validar que sea entero
					if v != float64(int64(v)) {
						errs = append(errs, ValidationError{Path: propPath, Err: "se esperaba un entero"})
					}
				case int, int32, int64:
					// ok
				default:
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo esperado integer, recibido %T", raw)})
				}

			case "object":
				// Debe ser map[string]any
				m, ok := raw.(map[string]any)
				if !ok {
					errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo esperado object, recibido %T", raw)})
					continue
				}
				// push frame hijo
				stack = append(stack, frame{
					path:                 propPath,
					schemaName:           p.Name,
					additionalProperties: falseIfZero(p.Properties, true /* default allow unless schema root indicates otherwise? */, f, p),
					props:                p.Properties,
					value:                m,
				})

			default:
				errs = append(errs, ValidationError{Path: propPath, Err: fmt.Sprintf("tipo no soportado en schema: %s", p.Type)})
			}
		}
	}

	return errs
}

// falseIfZero decide si permitir additionalProperties en objetos internos.
// Aquí tomamos la política: si el objeto anidado no declara nada explícito (tu modelo no tiene el flag en Property),
// usamos el mismo valor que en el nivel padre (f.additionalProperties). Puedes ajustar esta política.
func falseIfZero(_ []PropertyEntity, _default bool, parent frame, _prop PropertyEntity) bool {
	// No hay bandera en Property, así que heredamos del padre:
	return parent.additionalProperties
}

func joinPath(base, next string) string {
	if base == "" {
		return next
	}
	return base + "." + next
}

// ---------------------------
// Utilidades de formato
// ---------------------------

var emailRe = regexp.MustCompile(`^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`)

func isEmail(s string) bool {
	return emailRe.MatchString(s)
}

func isYYYYMMDD(s string) bool {
	if len(s) != 10 {
		return false
	}
	// Validación exacta YYYY-MM-DD
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

// ---------------------------
// Demo main
// ---------------------------

func main() {
	// Esquema de ejemplo (muy parecido a lo que has usado antes)
	schema := BodySchemaEntity{
		Name:                "CreateUserSchema",
		TypeSchema:          "object",
		AditionalProperties: true,
		Properties: []PropertyEntity{
			{
				Name:       "name",
				IsRequired: true,
				Type:       "string",
				MinLength:  1,
				MaxLength:  100,
			},
			{
				Name:       "email",
				IsRequired: true,
				Type:       "string",
				MinLength:  5,
				MaxLength:  255,
				Format:     "email",
			},
			{
				Name:       "birthdate",
				IsRequired: false,
				Type:       "string",
				Pattern:    `^\d{4}-\d{2}-\d{2}$`,
			},
			{
				Name:       "phone",
				IsRequired: true,
				Type:       "string",
				MinLength:  14,
				MaxLength:  14,
				Pattern:    `^\+52\s\d{10}$`,
			},
			{
				Name:       "address",
				IsRequired: true,
				Type:       "object",
				Properties: []PropertyEntity{
					{
						Name:       "street",
						IsRequired: true,
						Type:       "string",
						MinLength:  1,
						MaxLength:  200,
					},
				},
			},
		},
	}

	// Payload válido
	valid := map[string]any{
		"name":      "Rafael Zamora",
		"email":     "rafael.zamora@example.com",
		"birthdate": "2000-01-11",
		"phone":     "+53 5512345678",
		"address": map[string]any{
			"street": "Av. Universidad 3000, Ciudad de México",
		},
	}

	fmt.Println("== Validando payload válido ==")
	if errs := ValidateAgainstSchema(schema, valid); len(errs) == 0 {
		fmt.Println("OK ✅")
	} else {
		printErrs(errs)
	}

	// (Opcional) imprimir los JSON para verlos bonitos
	j1, _ := json.MarshalIndent(valid, "", "  ")
	fmt.Printf("\nPayload válido:\n%s\n", string(j1))
}

func printErrs(errs []ValidationError) {
	for _, e := range errs {
		fmt.Printf("- %s\n", e.String())
	}
}

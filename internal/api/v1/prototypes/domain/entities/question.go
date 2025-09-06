package entities

/*
{
    "request": {
        "method": "POST",
        "urlPath": "/api/users",
        "headers": {
            "Content-Type": {
                "contains": "application/json"
            },
            "Authorization": {
                "matches": "^Bearer\\s+[A-Za-z0-9\\-_]+\\.[A-Za-z0-9\\-_]+\\.[A-Za-z0-9\\-_]+$"
            }
        },
        "bodyPatterns": [
            {
                "matchesJsonSchema": {
                    "schemaVersion": "V202012",
                    "schema": {
                        "$schema": "https://json-schema.org/draft/2020-12/schema",
                        "type": "object",
                        "required": [
                            "name",
                            "email",
                            "birthdate",
                            "phone"
                        ],
                        "additionalProperties": false,
                        "properties": {
                            "name": {
                                "type": "string",
                                "minLength": 1
                            },
                            "email": {
                                "type": "string",
                                "format": "email",
                                "minLength": 3
                            },
                            "birthdate": {
                                "type": "string",
                                "format": "date"
                            },
                            "phone": {
                                "type": "string",
                                "pattern": "^\\+52\\s\\d{10}$"
                            }
                        }
                    }
                }
            }
        ]
    },
    "response": {
        "status": 201,
        "headers": {
            "Content-Type": "application/json"
        },
        "transformers": [
            "response-template"
        ],
        "transformerParameters": {
            "ignoreUndefinedVariables": true
        },
        "body": {
            "success": true,
            "status_code": 201,
            "data": {
                "id": "random.UUID",
                "name": "request.body.name",
                "email": "request.body.name",
                "birthdate": "request.body.birthdate",
                "phone": "+52 5513784057"
            }
        }
    }
}
*/

type PrototypeEntity struct {
	Name     string         `json:"name" binding:"required"`
	Request  RequestEntity  `json:"request" binding:"required"`
	Response ResponseEntity `json:"response" binding:"required"`
}

type RequestEntity struct {
	Method     string            `json:"method" binding:"required"`
	UrlPath    string            `json:"urlPath" binding:"required"`
	PathParams map[string]string `json:"path_params"`
	Headers    map[string]string `json:"headers"`
	BodySchema *BodySchemaEntity `json:"bodySchema"`

	Delay int `json:"delay"`
}

type BodySchemaEntity struct {
	Name                string           `json:"name" binding:"required"`
	TypeSchema          string           `json:"type_schema" binding:"required"`
	AditionalProperties bool             `json:"aditional_properties"`
	Properties          []PropertyEntity `json:"properties"`
}

type PropertyEntity struct {
	Name       string           `json:"name" binding:"required"`
	IsRequired bool             `json:"is_required"`
	Type       string           `json:"type" binding:"required"`
	MinLength  int32            `json:"min_length"`
	MaxLength  int32            `json:"max_length"`
	Format     string           `json:"format"`
	Pattern    string           `json:"pattern"`
	Properties []PropertyEntity `json:"properties"`
}

type ResponseEntity struct {
	Body map[string]any `json:"body"`
}

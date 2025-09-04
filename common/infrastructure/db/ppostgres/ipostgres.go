package ppostgres

import (
	"common/domain"
	"common/domain/criteria"
	"common/utils"
)

// --------------------------------
// INFRASTRUCTURE
// --------------------------------
// PostgresRepository
// --------------------------------

type IPostgresRepository[E domain.IEntity, M domain.IModel] interface {
	View(data []E)
	Save(role E) utils.Result[E]
	SearchAll() utils.Result[[]E]
	MatchingLow(cr criteria.Criteria, model *M) utils.Result[[]E]
	Delete(uuid string) error
	Search(uuid string) utils.Result[E]
	UpdateByFields(uuid string, fields map[string]interface{}) error
}

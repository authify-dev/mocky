package repositories

import (
	"common/domain/criteria"
	"common/domain/customctx"
	"common/utils"
	"context"
	"mocky/internal/db/mongo/prototypes"
)

type RepositoryPrototypes interface {
	Save(ctx context.Context, document prototypes.PrototypeModel) utils.Result[string]
	SaveWithID(ctx context.Context, id string, document prototypes.PrototypeModel) utils.Result[string]

	Update(ctx context.Context, entity prototypes.PrototypeModel) error
	UpdateFields(ctx context.Context, id string, updates map[string]interface{}) utils.Result[prototypes.PrototypeModel]

	Delete(ctx context.Context, id string) error

	Find(ctx context.Context, id string) utils.Result[prototypes.PrototypeModel]
	FindAll(ctx context.Context) utils.Result[[]prototypes.PrototypeListModel]

	Matching(cr criteria.Criteria, tableName string, offset int, limit int) utils.Result[[]prototypes.PrototypeListModel]
	GetByPath(cc *customctx.CustomContext, urlPath string, method string) utils.Result[prototypes.PrototypeModel]
	SaveOrUpdate(cc *customctx.CustomContext, document prototypes.PrototypeModel) utils.Result[string]
}

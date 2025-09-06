package prototypes

import (
	"common/domain/customctx"
	"common/domain/logger"
	ppmongo "common/infrastructure/db/ppmongo"
	"common/utils"
	"common/utils/cerrs"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// --------------------------------------
// Ropository of specific Entity
// --------------------------------------
type PrototypesMongoRepository struct {
	*ppmongo.MongoRepository[PrototypeModel, PrototypeListModel]
}

func NewPrototypesMongoRepository(uri string, dbName string, collectionName string) *PrototypesMongoRepository {
	return &PrototypesMongoRepository{
		MongoRepository: ppmongo.NewMongoRepository[PrototypeModel, PrototypeListModel](uri, dbName, collectionName),
	}
}

func (m *PrototypesMongoRepository) GetByPath(cc *customctx.CustomContext, urlPath string, method string) utils.Result[PrototypeModel] {
	entry := logger.FromContext(cc.Context())
	entry.Infof("Mongo GetByPath urlPath=%s", urlPath)

	var out PrototypeModel
	filter := bson.M{"request.urlpath": urlPath, "request.method": method}

	err := m.Collection.FindOne(cc.Context(), filter).Decode(&out)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.Result[PrototypeModel]{Err: cerrs.NewCustomError(http.StatusNotFound, "dont have prototype for this path: "+urlPath+" and method: "+method, "mongo.get_by_path")}
		}
		return utils.Result[PrototypeModel]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.get_by_path")}
	}

	return utils.Result[PrototypeModel]{Data: out}
}

func (m *PrototypesMongoRepository) SaveOrUpdate(cc *customctx.CustomContext, document PrototypeModel) utils.Result[string] {
	entry := logger.FromContext(cc.Context())
	entry.Infof("Mongo SaveOrUpdate document=%v", document)

	if document.Request.BodySchema != nil && document.Request.BodySchema.TypeSchema == "" {
		document.Request.BodySchema = nil
	}

	prototypeModel := m.GetByPath(cc, document.Request.UrlPath, document.Request.Method)

	// If the prototype does not exist, we save it
	if prototypeModel.Err != nil {
		document.CreatedAt = time.Now()
		document.UpdatedAt = time.Now()
		return m.MongoRepository.Save(cc.Context(), document)
	}

	// If the prototype exists, we update it

	err := m.MongoRepository.Delete(cc.Context(), prototypeModel.Data.ID)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.save_or_update")}
	}

	document.UpdatedAt = time.Now()

	newPrototype := m.MongoRepository.SaveWithID(cc.Context(), prototypeModel.Data.ID, document)
	if newPrototype.Err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, newPrototype.Err.Error(), "mongo.save_or_update")}
	}

	return utils.Result[string]{Data: newPrototype.Data}
}

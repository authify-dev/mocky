package mongo

import (
	"common/domain"
	"common/domain/criteria"
	"common/domain/logger"
	"common/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"common/utils/cerrs"
)

// --------------------------------------
// Ropository of specific Infra
// --------------------------------------
// MongoRepository es una implementación genérica de Repository usando generics.
type MongoRepository[T domain.IEntity, L domain.IEntity] struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

// NewMongoRepository crea una nueva instancia de MongoRepository.
func NewMongoRepository[T domain.IEntity, L domain.IEntity](uri string, dbName string, collectionName string) *MongoRepository[T, L] {
	// Context con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Configurar las opciones del cliente
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	// Crear referencia a la base de datos y la colección
	database := client.Database(dbName)
	collection := database.Collection(collectionName)

	return &MongoRepository[T, L]{
		Client:     client,
		Database:   database,
		Collection: collection,
	}
}

func (m *MongoRepository[T, L]) Save(ctx context.Context, document T) utils.Result[string] {
	result, err := m.Collection.InsertOne(ctx, document)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.save")}
	}

	// Convertir el InsertedID a primitive.ObjectID
	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, "el InsertedID no es un ObjectID válido", "mongo.save")}
	}

	// Retornar el ID como string
	return utils.Result[string]{Data: insertedID.Hex()}
}

// InsertDocumentWithID inserta un documento en MongoDB usando un UUID como _id
func (m *MongoRepository[T, L]) SaveWithID(ctx context.Context, id string, document T) utils.Result[string] {

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al convertir el id: %w", err).Error(), "mongo.save_with_id")}
	}

	// Convertir el documento a BSON
	data, err := bson.Marshal(document)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.save_with_id.marshal")}
	}

	// Deserializar el BSON a un mapa para poder modificarlo
	var docMap bson.M
	err = bson.Unmarshal(data, &docMap)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.save_with_id.unmarshal")}
	}

	// Establecer el _id con el valor proporcionado
	docMap["_id"] = _id

	// Insertar el documento modificado en la colección
	_, err = m.Collection.InsertOne(ctx, docMap)
	if err != nil {
		return utils.Result[string]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.save_with_id.insert_one")}
	}

	// Retornar el id como confirmación
	return utils.Result[string]{Data: _id.Hex()}
}

// Update actualiza un documento existente en la colección usando el id obtenido del entity.
func (m *MongoRepository[T, L]) Update(ctx context.Context, entity T) error {
	// Obtener el id desde el entity
	id := entity.GetID()

	// Convertir la entidad a BSON para poder actualizarla
	data, err := bson.Marshal(entity)
	if err != nil {
		return cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.update")
	}

	// Deserializar el BSON a un mapa para poder utilizarlo en la actualización
	var docMap bson.M
	err = bson.Unmarshal(data, &docMap)
	if err != nil {
		return cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.update")
	}

	// Crear el filtro basado en el _id
	filter := bson.M{"_id": id}

	// Realizar la actualización usando $set para reemplazar los campos existentes
	update := bson.M{"$set": docMap}
	result, err := m.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.update")
	}

	// Verificar que se haya encontrado y actualizado el documento
	if result.MatchedCount == 0 {
		return cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("no se encontró documento con id %s", id).Error(), "mongo.update")
	}

	return nil
}

// UpdateFields recibe el ID en hex string, aplica los updates y devuelve el objeto actualizado.
func (m *MongoRepository[T, L]) UpdateFields(ctx context.Context, id string, updates map[string]interface{}) utils.Result[T] {

	entry := logger.FromContext(ctx)
	entry.Info("Updating fields for document")

	var updated T

	// 1. Convertir hex string a ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		entry.Error("Error al convertir el ID a ObjectID", err)
		return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("id inválido (%s): %w", id, err).Error(), "mongo.update_fields")}
	}

	// 2. Validar campos a actualizar
	if len(updates) == 0 {
		entry.Error("No se proporcionaron campos para actualizar")
		return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("no se proporcionaron campos para actualizar").Error(), "mongo.update_fields")}
	}

	// 3. Preparar filtro y documento de actualización
	filter := bson.M{"_id": oid}
	update := bson.M{"$set": updates}

	// 4. Opciones: devolver el documento *después* de aplicar el update
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetUpsert(false) // evita crear uno nuevo si no existe

	// 5. Ejecutar FindOneAndUpdate y decodificar en `updated`
	err = m.Collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		entry.Error("Error al actualizar el documento", err)
		if err == mongo.ErrNoDocuments {
			entry.Errorf("no se encontró ningún documento con id %s", id)
			return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("no se encontró ningún documento con id %s", id).Error(), "mongo.update_fields")}
		}
		return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al devolver el documento actualizado: %w", err).Error(), "mongo.update_fields")}
	}

	return utils.Result[T]{Data: updated}
}

// Delete elimina un documento de la colección usando el id proporcionado.
func (m *MongoRepository[T, L]) Delete(ctx context.Context, id string) error {

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al convertir el id: %w", err).Error(), "mongo.delete")
	}

	// Crear el filtro usando _id
	filter := bson.M{"_id": _id}

	// Ejecutar la eliminación
	result, err := m.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al eliminar el documento: %w", err).Error(), "mongo.delete")
	}

	// Verificar que se haya eliminado algún documento
	if result.DeletedCount == 0 {
		return cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("no se encontró documento con id %s", id).Error(), "mongo.delete")
	}

	return nil
}

func (m *MongoRepository[T, L]) Find(ctx context.Context, id string) utils.Result[T] {
	var result T

	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al convertir el id: %w", err).Error(), "mongo.find")}
	}

	// Crear el filtro basado en el _id
	filter := bson.M{"_id": _id}

	// Ejecutar la búsqueda en la colección
	err = m.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("no se encontró documento con id %s", id).Error(), "mongo.find")}
		}
		return utils.Result[T]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al buscar el documento: %w", err).Error(), "mongo.find")}
	}

	return utils.Result[T]{Data: result}
}

func (m *MongoRepository[T, L]) FindAll(ctx context.Context) utils.Result[[]L] {

	var results []L
	cursor, err := m.Collection.Find(ctx, bson.M{})
	if err != nil {
		return utils.Result[[]L]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al buscar todos los documentos: %w", err).Error(), "mongo.find_all")}
	}

	if err := cursor.All(ctx, &results); err != nil {
		return utils.Result[[]L]{Err: cerrs.NewCustomError(http.StatusInternalServerError, fmt.Errorf("error al decodificar los documentos: %w", err).Error(), "mongo.find_all")}
	}

	return utils.Result[[]L]{Data: results}
}

func (m *MongoRepository[T, L]) Matching(cr criteria.Criteria, table_name string, offset int, limit int) utils.Result[[]L] {
	// Construir el filtro BSON basado en los criterios
	filter := bson.M{}

	// Recorrer los filtros para construir el filtro BSON
	for _, f := range cr.Filters.Get() {
		// Convertir el operador SQL a operador MongoDB
		mongoOperator := convertSQLOperatorToMongo(f.Operator, f.Value)

		// Aplicar el filtro al campo correspondiente
		if mongoOperator != nil {
			filter[string(f.Field)] = mongoOperator
		} else {
			// Para operadores simples como igualdad
			filter[string(f.Field)] = f.Value
		}
	}

	// Configurar opciones de búsqueda con paginación
	opts := options.Find()
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	// Ejecutar la consulta
	cursor, err := m.Collection.Find(context.Background(), filter, opts)
	if err != nil {
		return utils.Result[[]L]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.matching_low.find")}
	}
	defer cursor.Close(context.Background())

	// Decodificar los resultados
	var entities []L
	if err := cursor.All(context.Background(), &entities); err != nil {
		return utils.Result[[]L]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "mongo.matching_low.decode")}
	}

	return utils.Result[[]L]{Data: entities}
}

// convertSQLOperatorToMongo convierte operadores SQL a operadores MongoDB
func convertSQLOperatorToMongo(operator criteria.Operator, value interface{}) interface{} {
	switch operator {
	case criteria.OperatorEqual:
		return value
	case criteria.OperatorNotEqual:
		return bson.M{"$ne": value}
	case criteria.OperatorGreaterThan:
		return bson.M{"$gt": value}
	case criteria.OperatorGreaterEqual:
		return bson.M{"$gte": value}
	case criteria.OperatorLessThan:
		return bson.M{"$lt": value}
	case criteria.OperatorLessEqual:
		return bson.M{"$lte": value}
	case criteria.OperatorLike:
		// Convertir LIKE a regex de MongoDB
		if strValue, ok := value.(string); ok {
			return bson.M{"$regex": strValue, "$options": "i"}
		}
		return value
	case criteria.OperatorNotLike:
		// Convertir NOT LIKE a regex de MongoDB
		if strValue, ok := value.(string); ok {
			return bson.M{"$not": bson.M{"$regex": strValue, "$options": "i"}}
		}
		return bson.M{"$ne": value}
	case criteria.OperatorIn:
		// Para IN, el valor debe ser un slice
		return bson.M{"$in": value}
	case criteria.OperatorNotIn:
		// Para NOT IN, el valor debe ser un slice
		return bson.M{"$nin": value}
	default:
		return value
	}
}

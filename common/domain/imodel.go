package domain

import (
	"common/utils"
	"encoding/json"
	"net/http"

	"common/utils/cerrs"
)

// --------------------------------
// DOMAIN
// --------------------------------
// IModel
// --------------------------------
// Definimos una interfaz que represente a una entidad.
type IModel interface {
	GetID() string
	TableName() string
}

func ModelToEntity[E IEntity, M IModel](model IModel) utils.Result[E] {
	var result map[string]interface{}

	// Convertir el struct a JSON (bytes).
	data, err := json.Marshal(model)
	if err != nil {
		return utils.Result[E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "imodel.model_to_entity.Marshal")}
	}

	// Convertir los bytes JSON a un mapa.
	err = json.Unmarshal(data, &result)
	if err != nil {
		return utils.Result[E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "imodel.model_to_entity.Unmarshal")}
	}

	entity, err := FromJSON[E](result)
	if err != nil {
		return utils.Result[E]{Err: cerrs.NewCustomError(http.StatusInternalServerError, err.Error(), "imodel.model_to_entity.FromJSON")}
	}

	return utils.Result[E]{Data: entity}
}

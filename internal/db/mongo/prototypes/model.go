package prototypes

import "mocky/internal/api/v1/prototypes/domain/entities"

// Geolocalization es una implementación de Entity.
type PrototypeModel struct {
	ID      string                 `json:"id" bson:"_id,omitempty"`
	Request entities.RequestEntity `json:"request" bson:"request"`
	//Response entities.ResponseEntity `json:"response" bson:"response"`
}

func (g PrototypeModel) GetID() string {
	return g.ID
}

type PrototypeListModel struct {
	ID      string                 `json:"id" bson:"_id,omitempty"`
	Request entities.RequestEntity `json:"request" bson:"request"`
	//Response entities.ResponseEntity `json:"response" bson:"response"`
}

func (g PrototypeListModel) GetID() string {
	return g.ID
}

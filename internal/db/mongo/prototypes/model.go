package prototypes

import (
	"mocky/internal/api/v1/prototypes/domain/entities"
	"time"
)

// Geolocalization es una implementación de Entity.
type PrototypeModel struct {
	ID        string                  `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time               `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time               `json:"updatedAt" bson:"updatedAt"`
	Request   entities.RequestEntity  `json:"request" bson:"request"`
	Response  entities.ResponseEntity `json:"response" bson:"response"`
	Name      string                  `json:"name" bson:"name"`
}

func (g PrototypeModel) GetID() string {
	return g.ID
}

type PrototypeListModel struct {
	ID        string          `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt" bson:"updatedAt"`
	Request   RequestListView `json:"request" bson:"request"`
	Name      string          `json:"name" bson:"name"`
}

func (g PrototypeListModel) GetID() string {
	return g.ID
}

type RequestListView struct {
	Method  string `json:"method" binding:"required"`
	UrlPath string `json:"urlPath" binding:"required"`
}

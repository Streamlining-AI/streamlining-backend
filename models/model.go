package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type Model struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `json:"Name" validate:"required,min=6"`
	ImageId    string             `json:"ImageId" validate:"email,required"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	Input      string             `json:"Input"`
	Url        string             `json:"Url"`
	PredictUrl string             `json:"PredictUrl"`
}

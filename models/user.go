package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Password      *string            `json:"Password" validate:"required,min=6"`
	Email         *string            `json:"email" validate:"email,required"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
}

type UserGithub struct {
	ID         primitive.ObjectID `bson:"_id"`
	Username   string             `json:"username" validate:"required"`
	Created_at time.Time          `json:"created_at"`
	Updated_at time.Time          `json:"updated_at"`
	User_id    string             `json:"user_id"`
}

type GithubRequestBody struct {
	Code     string `json:"Code"`
	Username string `json:"Username"`
	UserId   string `json:"UserId"`
}

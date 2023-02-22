package schemas

import (
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ObjectID = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "BSON",
	Description: "The `bson` scalar type represents a BSON Object.",
	// Serialize serializes `bson.ObjectId` to string.
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case primitive.ObjectID:
			return value.Hex()
		case *primitive.ObjectID:
			v := *value
			return v.Hex()
		default:
			return nil
		}
	},
	// ParseValue parses GraphQL variables from `string` to `bson.ObjectId`.
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case string:
			id, _ := primitive.ObjectIDFromHex(value)
			return id
		default:
			return nil
		}
	}})

var ModelDataType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "ModelData",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: ObjectID,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: graphql.String,
			},
			"isVisible": &graphql.Field{
				Type: graphql.Boolean,
			},
			"githubURL": &graphql.Field{
				Type: graphql.String,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"predictRecordCount": &graphql.Field{
				Type: graphql.Float,
			},
			"createdAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"updatedAt": &graphql.Field{
				Type: graphql.DateTime,
			},
			"userID": &graphql.Field{
				Type: ObjectID,
			},
			"outputType": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

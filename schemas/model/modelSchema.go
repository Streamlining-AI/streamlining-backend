package schemas

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/graphql-go/graphql"
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			/* Get (read) single product by id
			   http://localhost:8080/product?query={product(id:1){name,info,price}}
			*/
			"model": &graphql.Field{
				Type:        ModelDataType,
				Description: "Get model by id",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if ok {
						// Find product Query DB===========================================
						// return model, nil
						fmt.Print(id)
					}
					return nil, nil
				},
			},

			/* Get (read) product list
			   http://localhost:8080/product?query={list{id,name,info,price}}
			*/
			"list": &graphql.Field{
				Type:        graphql.NewList(ModelDataType),
				Description: "Get model list",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

					// Query from DB ====================================================
					// return Models, nil ===============================
					return nil, nil
				},
			},
		},
	})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		/* Create new product item
		http://localhost:8080/product?query=mutation+_{create(name:"Inca Kola",info:"Inca Kola is a soft drink that was created in Peru in 1935 by British immigrant Joseph Robinson Lindley using lemon verbena (wiki)",price:1.99){id,name,info,price}}
		*/
		"create": &graphql.Field{
			Type:        ModelDataType,
			Description: "Create new model",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: ObjectID,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"isVisible": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
				},
				"githubURL": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"predictRecordCount": &graphql.ArgumentConfig{
					Type: graphql.Float,
				},
				"createdAt": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"updatedAt": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"userID": &graphql.ArgumentConfig{
					Type: ObjectID,
				},
				"outputType": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				rand.Seed(time.Now().UnixNano())
				// Query and Insert
				// model := Models.MLModel{
				// 	ModelID:    params.Args["id"].(int),
				// 	Name:       params.Args["name"].(string),
				// 	ImageID:    params.Args["imageid"].(string),
				// 	Input:      params.Args["input"].(string),
				// 	URL:        params.Args["url"].(string),
				// 	PredictURL: params.Args["predicturl"].(string),
				// }
				// models = append(models, model)
				// Return data, nil
				return nil, nil
			},
		},

		/* Update product by id
		   http://localhost:8080/product?query=mutation+_{update(id:1,price:3.95){id,name,info,price}}
		*/
		"update": &graphql.Field{
			Type:        ModelDataType,
			Description: "Update model by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: ObjectID,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"isVisible": &graphql.ArgumentConfig{
					Type: graphql.Boolean,
				},
				"githubURL": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"description": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"predictRecordCount": &graphql.ArgumentConfig{
					Type: graphql.Float,
				},
				"createdAt": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"updatedAt": &graphql.ArgumentConfig{
					Type: graphql.DateTime,
				},
				"userID": &graphql.ArgumentConfig{
					Type: ObjectID,
				},
				"outputType": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

				// Query and update
				// id, _ := params.Args["id"].(int)
				// name, nameOk := params.Args["name"].(string)
				// imageid, imageidOk := params.Args["imageid"].(string)
				// input, inputOk := params.Args["input"].(string)
				// url, urlOk := params.Args["url"].(string)
				// predicturl, predicturlOk := params.Args["predicturl"].(string)

				// Return data, nil
				return nil, nil
			},
		},

		/* Delete product by id
		   http://localhost:8080/product?query=mutation+_{delete(id:1){id,name,info,price}}
		*/
		"delete": &graphql.Field{
			Type:        ModelDataType,
			Description: "Delete product by id",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				// Delete from DB===========================================================
				// Retirn data,nil
				return nil, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

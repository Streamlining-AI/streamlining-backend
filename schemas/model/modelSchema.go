package schemas

// import (
// 	controller "github.com/Streamlining-AI/streamlining-backend/controllers"
// 	"github.com/Streamlining-AI/streamlining-backend/models"
// 	"github.com/graphql-go/graphql"
// )

// var queryType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "Query",
// 		Fields: graphql.Fields{
// 			/* Get (read) single product by id
// 			   http://localhost:8080/product?query={product(id:1){name,info,price}}
// 			*/
// 			"model": &graphql.Field{
// 				Type:        ModelDataType,
// 				Description: "Get model by id",
// 				Args: graphql.FieldConfigArgument{
// 					"model_id": &graphql.ArgumentConfig{
// 						Type: graphql.String,
// 					},
// 				},
// 				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
// 					id, ok := p.Args["model_id"].(string)
// 					var Model models.ModelData
// 					var errr error
// 					if ok {
// 						model, err := controller.GetModelByID(id)
// 						Model = model
// 						errr = err
// 					}
// 					return Model, errr
// 				},
// 			},

// 			/* Get (read) product list
// 			   http://localhost:8080/product?query={list{id,name,info,price}}
// 			*/
// 			"list": &graphql.Field{
// 				Type:        graphql.NewList(ModelDataType),
// 				Description: "Get model list",
// 				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

// 					models := controller.GetAllModel1()
// 					// Query from DB ====================================================
// 					// return Models, nil ===============================
// 					return models, nil
// 				},
// 			},
// 		},
// 	})

// var mutationType = graphql.NewObject(graphql.ObjectConfig{
// 	Name: "Mutation",
// 	Fields: graphql.Fields{
// 		/* Create new product item
// 		http://localhost:8080/product?query=mutation+_{create(name:"Inca Kola",info:"Inca Kola is a soft drink that was created in Peru in 1935 by British immigrant Joseph Robinson Lindley using lemon verbena (wiki)",price:1.99){id,name,info,price}}
// 		*/
// 		"create": &graphql.Field{
// 			Type:        ModelDataType,
// 			Description: "Create new model",
// 			Args: graphql.FieldConfigArgument{
// 				"name": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"type": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"is_visible": &graphql.ArgumentConfig{
// 					Type: graphql.Boolean,
// 				},
// 				"github_url": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"description": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"user_id": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"output_type": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"github_code": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 			},
// 			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
// 				var Model models.ModelDataTransfer
// 				Model.Name = params.Args["name"].(string)
// 				Model.Type = params.Args["type"].(string)
// 				Model.IsVisible = params.Args["is_visible"].(bool)
// 				Model.GithubURL = params.Args["github_url"].(string)
// 				Model.Description = params.Args["description"].(string)
// 				Model.OutputType = params.Args["output_type"].(string)
// 				Model.UserID = params.Args["user_id"].(string)
// 				// GithubCode := params.Args["github_code"].(string)
// 				// ModelData := controller.HandlerUpload1(Model, GithubCode)

// 				return Model, nil
// 			},
// 		},

// 		/* Update product by id
// 		   http://localhost:8080/product?query=mutation+_{update(id:1,price:3.95){id,name,info,price}}
// 		*/
// 		"update": &graphql.Field{
// 			Type:        ModelDataType,
// 			Description: "Update model by id",
// 			Args: graphql.FieldConfigArgument{
// 				"id": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"name": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"type": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"isVisible": &graphql.ArgumentConfig{
// 					Type: graphql.Boolean,
// 				},
// 				"githubURL": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"description": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 				"predictRecordCount": &graphql.ArgumentConfig{
// 					Type: graphql.Float,
// 				},
// 				"createdAt": &graphql.ArgumentConfig{
// 					Type: graphql.DateTime,
// 				},
// 				"updatedAt": &graphql.ArgumentConfig{
// 					Type: graphql.DateTime,
// 				},
// 				"userID": &graphql.ArgumentConfig{
// 					Type: ObjectID,
// 				},
// 				"outputType": &graphql.ArgumentConfig{
// 					Type: graphql.String,
// 				},
// 			},
// 			Resolve: func(params graphql.ResolveParams) (interface{}, error) {

// 				// Query and update
// 				// id, _ := params.Args["id"].(int)
// 				// name, nameOk := params.Args["name"].(string)
// 				// imageid, imageidOk := params.Args["imageid"].(string)
// 				// input, inputOk := params.Args["input"].(string)
// 				// url, urlOk := params.Args["url"].(string)
// 				// predicturl, predicturlOk := params.Args["predicturl"].(string)

// 				// Return data, nil
// 				return nil, nil
// 			},
// 		},

// 		/* Delete product by id
// 		   http://localhost:8080/product?query=mutation+_{delete(id:1){id,name,info,price}}
// 		*/
// 		"delete": &graphql.Field{
// 			Type:        ModelDataType,
// 			Description: "Delete product by id",
// 			Args: graphql.FieldConfigArgument{
// 				"id": &graphql.ArgumentConfig{
// 					Type: graphql.NewNonNull(graphql.Int),
// 				},
// 			},
// 			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
// 				// Delete from DB===========================================================
// 				// Retirn data,nil
// 				return nil, nil
// 			},
// 		},
// 	},
// })

// var Schema, _ = graphql.NewSchema(
// 	graphql.SchemaConfig{
// 		Query:    queryType,
// 		Mutation: mutationType,
// 	},
// )

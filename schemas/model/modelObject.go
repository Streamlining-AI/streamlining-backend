package schemas

// import (
// 	"github.com/graphql-go/graphql"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// var ObjectID = graphql.NewScalar(graphql.ScalarConfig{
// 	Name:        "BSON",
// 	Description: "The `bson` scalar type represents a BSON Object.",
// 	// Serialize serializes `bson.ObjectId` to string.
// 	Serialize: func(value interface{}) interface{} {
// 		switch value := value.(type) {
// 		case primitive.ObjectID:
// 			return value.Hex()
// 		case *primitive.ObjectID:
// 			v := *value
// 			return v.Hex()
// 		default:
// 			return nil
// 		}
// 	},
// 	// ParseValue parses GraphQL variables from `string` to `bson.ObjectId`.
// 	ParseValue: func(value interface{}) interface{} {
// 		switch value := value.(type) {
// 		case string:
// 			id, _ := primitive.ObjectIDFromHex(value)
// 			return id
// 		default:
// 			return nil
// 		}
// 	}})

// var ModelDataType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelData",
// 		Fields: graphql.Fields{
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"name": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"type": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"is_visible": &graphql.Field{
// 				Type: graphql.Boolean,
// 			},
// 			"github_url": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"description": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"predict_record_count": &graphql.Field{
// 				Type: graphql.Float,
// 			},
// 			"created_at": &graphql.Field{
// 				Type: graphql.DateTime,
// 			},
// 			"updated_at": &graphql.Field{
// 				Type: graphql.DateTime,
// 			},
// 			"user_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"output_type": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelImageType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelImage",
// 		Fields: graphql.Fields{
// 			"image_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelPodType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelPod",
// 		Fields: graphql.Fields{
// 			"pod_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"predict_url": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"image_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelReportType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelReport",
// 		Fields: graphql.Fields{
// 			"model_report_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"description": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"created_at": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"user_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelInputDetailType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelInputDetail",
// 		Fields: graphql.Fields{
// 			"model_input_detail_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"name": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"type": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"description": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"default": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"max": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"min": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelInputType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelInput",
// 		Fields: graphql.Fields{
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"input_detail": &graphql.Field{
// 				Type: graphql.NewList(ModelInputDetailType),
// 			},
// 		},
// 	},
// )

// var ModelOutputType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelOutput",
// 		Fields: graphql.Fields{
// 			"model_output_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"output": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"created_at": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"model_input_data": &graphql.Field{
// 				Type: ModelInputDataType,
// 			},
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var ModelInputDataType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "ModelInputWithData",
// 		Fields: graphql.Fields{
// 			"model_input_data_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"data_inputs": &graphql.Field{
// 				Type: graphql.NewList(DataInputType),
// 			},
// 			"image_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"model_id": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// var DataInputType = graphql.NewObject(
// 	graphql.ObjectConfig{
// 		Name: "DataInput",
// 		Fields: graphql.Fields{
// 			"data": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"name": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 			"type": &graphql.Field{
// 				Type: graphql.String,
// 			},
// 		},
// 	},
// )

// //

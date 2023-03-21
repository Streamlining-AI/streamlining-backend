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

type PredictBody struct {
	Id    string `json:"id"`
	Input string `json:"input"`
}

type Message struct {
	Url   string `json:"url"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Input string `json:"input"`
}

type ModelDataTransfer struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	IsVisible     bool   `json:"is_visible"`
	GithubURL     string `json:"github_url"`
	UserID        string `json:"user_id"`
	OutputType    string `json:"output_type"`
	GithubCode    string `json:"github_code"`
	Model_Version string `json:"model_version"`
	Banner        string `json:"banner_url"`
}

type ModelData struct {
	ModelID            primitive.ObjectID `bson:"model_id" json:"model_id"`
	Name               string             `bson:"name" json:"name"`
	Type               string             `bson:"type" json:"type"`
	IsVisible          bool               `bson:"is_visible" json:"is_visible"`
	GithubURL          string             `bson:"github_url" json:"github_url"`
	Description        string             `bson:"description" json:"description"`
	PredictRecordCount int                `bson:"predict_record_count" json:"predict_record_count"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
	UserID             primitive.ObjectID `bson:"user_id" json:"user_id"`
	OutputType         string             `bson:"output_type" json:"output_type"`
	Banner             string             `bson:"banner_url" json:"banner_url"`
}

type ModelImage struct {
	ImageID       primitive.ObjectID `bson:"image_id" json:"image_id"`
	DockerImageID string             `bson:"docker_image_id" json:"docker_image_id"`
	ModelID       primitive.ObjectID `bson:"model_id" json:"model_id"`
	ModelVersion  string             `bson:"model_version" json:"model_version"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

type ModelPod struct {
	PodID        primitive.ObjectID `bson:"pod_id" json:"pod_id"`
	PodURL       string             `bson:"pod_url" json:"pod_url"`
	PredictURL   string             `bson:"predict_url" json:"predict_url"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	RecentUsedAt time.Time          `bson:"recent_used_at" json:"recent_used_at"`
	ImageID      primitive.ObjectID `bson:"image_id" json:"image_id"`
}

type ModelReport struct {
	ModelReportID primitive.ObjectID `bson:"model_report_id"`
	Description   string             `bson:"description" json:"description"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	ModelID       primitive.ObjectID `bson:"model_id" json:"model_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
}

type ModelInput struct {
	ModelID       primitive.ObjectID `bson:"model_id" json:"model_id"`
	DockerImageID string             `bson:"docker_image_id" json:"docker_image_id"`
	InputDetail   []ModelInputDetail `bson:"input_detail" json:"input_detail"`
}

type ModelInputDetail struct {
	ModelInputDetailID primitive.ObjectID `bson:"model_input_detail_id"`
	Name               string             `bson:"name" json:"name"`
	Type               string             `bson:"type" json:"type"`
	Description        string             `bson:"description" json:"description"`
	Default            interface{}        `bson:"default" json:"default"`
	Optional           Optional           `bson:"optional" json:"optional"`
}

type ModelOutputData struct {
	ModelOutputID  primitive.ObjectID `bson:"model_output_id"`
	Output         string             `bson:"output" json:"output"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	ModelInputData ModelInputData     `bson:"model_input_data" json:"model_input_data"`
	ModelID        primitive.ObjectID `bson:"model_id" bson:"model_id"`
}

type ModelInputData struct {
	DataInputs    []DataInput `bson:"data_inputs" json:"data_inputs"`
	DockerImageID string      `bson:"docker_image_id" json:"docker_image_id"`
}

type ModelConfig struct {
	Input  []Input `json:"input"`
	Output Output  `json:"output"`
}

type Output struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Input struct {
	Name        string      `bson:"name" json:"name"`
	Type        string      `bson:"type" json:"type"`
	Description string      `bson:"description" json:"description"`
	Default     interface{} `bson:"default" json:"default"`
	Optional    Optional    `bson:"optional" json:"optional"`
}

type Optional struct {
	MaxLength int           `bson:"max_length" json:"max_length"`
	MinLength int           `bson:"min_length" json:"min_length"`
	Ge        int           `bson:"ge" json:"ge"`
	Le        int           `bson:"le" json:"le"`
	Choices   []interface{} `bson:"choices" json:"choices"`
}

type DataInput struct {
	ModelInputDetailID string      `bson:"model_input_detail_id" json:"model_input_detail_id"`
	Name               string      `bson:"name" json:"name"`
	Type               string      `bson:"type" json:"type"`
	Data               interface{} `bson:"data" json:"data"`
}

type ModelInputDataTransfer struct {
	DataInputs    []DataInput `bson:"data_inputs" json:"data_inputs"`
	ModelID       string      `bson:"model_id" json:"model_id"`
	DockerImageID string      `bson:"docker_image_id" json:"docker_image_id"`
}

type ModelReportRequest struct {
	ModelID     string `json:"model_id"`
	UserID      string `json:"user_id"`
	Description string `json:"description"`
}

type ModelTransfer struct {
	ModelID            primitive.ObjectID `json:"model_id"`
	Name               string             `json:"name"`
	Type               string             `json:"type"`
	GithubURL          string             `json:"github_url"`
	Description        string             `json:"description"`
	PredictRecordCount int                `json:"predict_record_count"`
	CreatedAt          time.Time          `json:"created_at"`
	OutputType         string             `json:"output_type"`
	DockerImageID      []string           `json:"docker_image_id"`
	InputDetail        []ModelInputDetail `json:"input_detail"`
}

type ModelIDAndDockerImageID struct {
	ModelID       string `json:"model_id"`
	DockerImageID string `json:"docker_image_id"`
}

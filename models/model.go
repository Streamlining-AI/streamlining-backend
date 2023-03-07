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
	Name        string `json:"name" validate:"required,min=6"`
	Type        string `json:"type" validate:"email,required"`
	IsVisible   bool   `json:"is_visible"`
	GithubURL   string `json:"github_url"`
	Description string `json:"description"`
	UserID      string `json:"user_id"`
	OutputType  string `json:"output_type"`
	GithubCode  string `json:"github_code"`
}

type ModelData struct {
	ModelID            primitive.ObjectID `bson:"model_id"`
	Name               string             `json:"name" validate:"required,min=6"`
	Type               string             `json:"type" validate:"email,required"`
	IsVisible          bool               `json:"is_visible"`
	GithubURL          string             `json:"github_url"`
	Description        string             `json:"description"`
	PredictRecordCount int                `json:"predict_record_count"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	UserID             primitive.ObjectID `bson:"user_id"`
	OutputType         string             `json:"output_type"`
}

type ModelImage struct {
	ImageID           primitive.ObjectID `bson:"image_id"`
	DockerImageID     string             `bson:"docker_image_id"`
	DockeyRegistryURL string             `bson:"docker_registry_url"`
	ModelID           primitive.ObjectID `bson:"model_id"`
}

type ModelPod struct {
	PodID        primitive.ObjectID `bson:"pod_id"`
	PodURL       string             `bson:"pod_url"`
	PredictURL   string             `bson:"predict_url"`
	CreatedAt    time.Time          `bson:"created_at"`
	RecentUsedAt time.Time          `bson:"recent_used_at"`
	ImageID      primitive.ObjectID `bson:"image_id"`
}

type ModelReport struct {
	ModelReportID primitive.ObjectID `bson:"model_report_id"`
	Description   string             `json:"description"`
	CreatedAt     time.Time          `json:"created_at"`
	ModelID       primitive.ObjectID `bson:"model_id"`
	UserID        primitive.ObjectID `bson:"user_id"`
}

type ModelInput struct {
	ModelID       primitive.ObjectID `bson:"model_id"`
	DockerImageID string             `json:"docker_image_id"`
	InputDetail   []ModelInputDetail `json:"input_detail"`
}

type ModelInputDetail struct {
	ModelInputDetailID primitive.ObjectID `bson:"model_input_detail_id"`
	Name               string             `json:"name"`
	Type               string             `json:"type"`
	Description        string             `json:"description"`
	Default            interface{}        `json:"default"`
	Optional           Optional           `json:"optional"`
}

type ModelOutputData struct {
	ModelOutputID  primitive.ObjectID `bson:"model_output_id"`
	Output         string             `bson:"output"`
	CreatedAt      time.Time          `bson:"created_at"`
	ModelInputData ModelInputData     `bson:"model_input_data"`
	ModelID        primitive.ObjectID `bson:"model_id"`
}

type ModelInputData struct {
	DataInputs    []DataInput `json:"data_inputs"`
	DockerImageID string      `json:"docker_image_id"`
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
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default"`
	Optional    Optional    `json:"optional"`
}

type Optional struct {
	MaxLength int           `json:"max_length"`
	MinLength int           `json:"min_length"`
	Ge        int           `json:"ge"`
	Le        int           `json:"le"`
	Choices   []interface{} `json:"choices"`
}

type DataInput struct {
	Data               interface{} `json:"data"`
	ModelInputDetailID string      `json:"model_input_detail_id"`
	Type               string      `json:"type"`
}

type ModelInputDataTransfer struct {
	DataInputs    []DataInput `json:"data_inputs"`
	ModelID       string      `json:"model_id"`
	DockerImageID string      `json:"docker_image_id"`
}

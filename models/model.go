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

type ModelData struct {
	ModelID            primitive.ObjectID `bson:"model_id"`
	Name               string             `json:"name" validate:"required,min=6"`
	Type               string             `json:"type" validate:"email,required"`
	IsVisible          bool               `json:"is_visible"`
	GithubURL          string             `json:"github_url"`
	Description        string             `json:"description"`
	PredictRecordCount float32            `json:"predict_record_count"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	UserID             primitive.ObjectID `bson:"user_id"`
	OutputType         string             `json:"output_type"`
}

type ModelImage struct {
	ImageID            primitive.ObjectID `bson:"image_id"`
	DockerImageID      string             `json:"docker_image_id"`
	DockerRegistryName string             `json:"docker_registry_name"`
	DockeyRegistryURL  string             `json:"docker_registry_url"`
	ModelID            primitive.ObjectID `bson:"model_id"`
}

type ModelPod struct {
	PodID        primitive.ObjectID `bson:"pod_id"`
	PodURL       string             `json:"pod_url"`
	PredictURL   string             `json:"predict_url"`
	CreatedAt    time.Time          `json:"created_at"`
	RecentUsedAt time.Time          `json:"recent_used_at"`
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
	ModelID     primitive.ObjectID `bson:"model_id"`
	InputDetail []ModelInputDetail `json:"input_detail"`
}

type ModelInputDetail struct {
	ModelInputDetailID primitive.ObjectID `bson:"model_input_detail_id"`
	Name               string             `json:"name"`
	Type               string             `json:"type"`
	Description        string             `json:"description"`
	Default            string             `json:"default"`
	Max                string             `json:"max"`
	Min                string             `json:"min"`
}

type ModelOutputData struct {
	ModelOutputID  primitive.ObjectID `bson:"model_output_id"`
	Output         string             `json:"output"`
	CreatedAt      time.Time          `json:"created_at"`
	ModelInputData []ModelInputData   `json:"model_input_data"`
}

type ModelInputData struct {
	ModelInputDataID     primitive.ObjectID     `bson:"model_input_data_id"`
	ModelInputDataDetail []ModelInputDataDetail `json:"model_input_data_detail"`
}

type ModelInputDataDetail struct {
	Data string `json:"data"`
	Name string `json:"name"`
	Type string `json:"type"`
}

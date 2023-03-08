package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	gohttp "net/http"
	"os"
	"strings"
	"time"

	"github.com/Streamlining-AI/streamlining-backend/database"
	helper "github.com/Streamlining-AI/streamlining-backend/helpers"
	"github.com/Streamlining-AI/streamlining-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var modelCollection *mongo.Collection = database.OpenCollection(database.Client, "model")
var modelCollectionImage *mongo.Collection = database.OpenCollection(database.Client, "model_image")
var modelCollectionPod *mongo.Collection = database.OpenCollection(database.Client, "model_pod")
var modelCollectionReport *mongo.Collection = database.OpenCollection(database.Client, "model_report")
var modelCollectionInputDetail *mongo.Collection = database.OpenCollection(database.Client, "model_input_detail")
var modelCollectionOutput *mongo.Collection = database.OpenCollection(database.Client, "model_output")

func HandlerUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var reqModel models.ModelDataTransfer
		defer cancel()
		if err := c.BindJSON(&reqModel); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var model models.ModelData
		model.ModelID = primitive.NewObjectID()
		model.Name = reqModel.Name
		model.Type = reqModel.Type
		model.IsVisible = reqModel.IsVisible
		model.GithubURL = reqModel.GithubURL
		model.Description = reqModel.Description
		model.PredictRecordCount = 0
		model.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.UserID, _ = primitive.ObjectIDFromHex(reqModel.UserID)
		_, err := modelCollection.InsertOne(ctx, model)

		defer cancel()
		if err != nil {
			fmt.Print(err)
		}
		// githubCode := reqModel.GithubCode

		// claims, _ := helper.DecodeToken(githubCode)
		dir := "repos/" + model.Name

		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			// Auth: &http.BasicAuth{
			// 	Username: "arbruzaz",
			// 	Password: claims.AccessToken,
			// },
			URL:               model.GithubURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		// ImageID, DockerURL, DockerImageID, err := HandlerDeployDocker(dir, model.Name, model.ModelID)
		if err != nil {
			fmt.Print(err)
		}

		DockerImageID := "TestImage"
		err = HandlerConfig(dir, model.ModelID, DockerImageID)
		if err != nil {
			fmt.Print(err)
		}

		// err = HandlerDeployKube(DockerURL, ImageID, model.Name)

		if err != nil {
			fmt.Print(err)
		}

		c.JSON(200, gin.H{"message": "Clone Successful"})
	}
}

func HandlerConfig(dir string, modelID primitive.ObjectID, dockerImageID string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Read Config file
	fileContent, err := os.Open(dir + "/config.json")
	defer cancel()
	if err != nil {
		return err
	}

	fmt.Println("The File is opened successfully...")

	defer fileContent.Close()

	byteResult, _ := ioutil.ReadAll(fileContent)

	var payload models.ModelConfig
	json.Unmarshal(byteResult, &payload)

	// Create And Insert Inputs Detail Config
	InputDetail := CreateInputDetail(payload.Input)

	var modelInputs models.ModelInput
	modelInputs.ModelID = modelID
	modelInputs.DockerImageID = dockerImageID
	modelInputs.InputDetail = InputDetail
	// Insert Input Detail to DB
	_, err = modelCollectionInputDetail.InsertOne(ctx, modelInputs)

	if err != nil {
		return err
	}
	// Update Output Detail to Model Collection
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "output_type", Value: payload.Output.Type}}}}
	_, err = modelCollection.UpdateOne(ctx, bson.M{"model_id": modelID}, update)

	if err != nil {
		return err
	}

	return nil
}

func CreateInputDetail(inputConfig []models.Input) []models.ModelInputDetail {
	var modelInputDetailS []models.ModelInputDetail

	for i := 0; i < len(inputConfig); i++ {
		var modelInputDetail models.ModelInputDetail
		modelInputDetail.ModelInputDetailID = primitive.NewObjectID()
		modelInputDetail.Name = inputConfig[i].Name
		modelInputDetail.Type = inputConfig[i].Type
		modelInputDetail.Description = inputConfig[i].Description
		modelInputDetail.Default = inputConfig[i].Default
		modelInputDetail.Optional = inputConfig[i].Optional
		modelInputDetailS = append(modelInputDetailS, modelInputDetail)
	}
	return modelInputDetailS
}

func HandlerDeployDocker(dir string, modelName string, modelID primitive.ObjectID) (primitive.ObjectID, string, string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	DockerImageID, DockerURL := helper.PushToDocker(dir, modelName)

	var modelImage models.ModelImage
	modelImage.ImageID = primitive.NewObjectID()
	modelImage.DockerImageID = DockerImageID
	modelImage.DockeyRegistryURL = DockerURL
	modelImage.ModelID = modelID
	_, err := modelCollectionImage.InsertOne(ctx, modelImage)

	defer cancel()
	if err != nil {
		return modelImage.ImageID, DockerURL, modelImage.DockerImageID, err
	}

	return modelImage.ImageID, DockerURL, modelImage.DockerImageID, nil
}

func HandlerDeployKube(DockerURL string, ImageID primitive.ObjectID, modelName string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	helper.CreateDeploy(DockerURL, modelName)
	helper.CreateService(modelName)
	PredictURL, PodURL := helper.DeployKube(modelName)

	var modelPod models.ModelPod
	modelPod.PodID = primitive.NewObjectID()
	modelPod.PodURL = PodURL
	modelPod.PredictURL = PredictURL
	modelPod.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	modelPod.ImageID = ImageID

	_, err := modelCollectionPod.InsertOne(ctx, modelPod)

	defer cancel()
	if err != nil {
		return err
	}
	return nil
}

func GetAllModel() gin.HandlerFunc {
	return func(c *gin.Context) {
		var foundModels []models.ModelData
		findOptions := options.Find()

		cur, err := modelCollection.Find(context.TODO(), bson.D{{}}, findOptions)
		if err != nil {
			println(err)
			return
		}

		for cur.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem models.ModelData
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}

			foundModels = append(foundModels, elem)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		//Close the cursor once finished
		cur.Close(context.TODO())

		fmt.Printf("Found multiple documents: %+v\n", foundModels)
		c.JSON(200, foundModels)
		return
	}
}

func GetModelByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		id := c.Param("model_id")
		modelID, _ := primitive.ObjectIDFromHex(id)
		var foundModel models.ModelData

		err := modelCollection.FindOne(ctx, bson.M{"model_id": modelID}).Decode(&foundModel)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Model ID is invalid"})
			return
		}

		// var modelInput models.ModelInput

		// err = modelCollectionInputDetail.FindOne(ctx, bson.M{"model_id": modelID, "docker_image_id": inputData.DockerImageID}).Decode(&modelInput)
		// defer cancel()
		// if err != nil {
		// 	c.JSON(400, gin.H{"error": err.Error()})
		// 	return
		// }

		c.JSON(200, foundModel)
		return
	}
}

func HandlerPredict() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var inputData models.ModelInputDataTransfer

		defer cancel()
		if err := c.BindJSON(&inputData); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		modelID, _ := primitive.ObjectIDFromHex(inputData.ModelID)
		var model models.ModelData
		err := modelCollection.FindOne(ctx, bson.M{"model_id": modelID}).Decode(&model)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var modelInput models.ModelInput

		err = modelCollectionInputDetail.FindOne(ctx, bson.M{"model_id": modelID, "docker_image_id": inputData.DockerImageID}).Decode(&modelInput)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		predictBody := map[string]map[string]interface{}{}
		predictBody["input"] = map[string]interface{}{}
		for i := 0; i < len(inputData.DataInputs); i++ {
			inputDetailID, _ := primitive.ObjectIDFromHex(inputData.DataInputs[i].ModelInputDetailID)
			for j := 0; j < len(modelInput.InputDetail); j++ {
				if inputDetailID == modelInput.InputDetail[j].ModelInputDetailID {
					if inputData.DataInputs[i].Type != modelInput.InputDetail[j].Type {
						c.JSON(400, gin.H{"error": "Wrong Input Type"})
						return
					}
					predictBody["input"][modelInput.InputDetail[j].Name] = inputData.DataInputs[i].Data
					break
				}
			}
		}

		var modelImage models.ModelImage
		err = modelCollectionImage.FindOne(ctx, bson.M{"docker_image_id": inputData.DockerImageID}).Decode(&modelImage)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var modelPod models.ModelPod
		err = modelCollectionPod.FindOne(ctx, bson.M{"image_id": modelImage.ImageID}).Decode(&modelPod)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var modelInputData models.ModelInputData
		modelInputData.DockerImageID = inputData.DockerImageID
		modelInputData.DataInputs = inputData.DataInputs

		var modelOutputData models.ModelOutputData
		modelOutputData.ModelOutputID = primitive.NewObjectID()
		modelOutputData.Output = "Output1"
		modelOutputData.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		modelOutputData.ModelID = modelID
		modelOutputData.ModelInputData = modelInputData

		result, err := modelCollectionOutput.InsertOne(ctx, modelOutputData)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "predict_record_count", Value: model.PredictRecordCount + 1}}}}
		_, err = modelCollection.UpdateOne(ctx, bson.M{"model_id": modelID}, update)
		// requestJSON, _ := json.Marshal(predictBody)
		// req, reqerr := gohttp.NewRequest(
		// 	"POST",
		// 	modelPod.PredictURL,
		// 	bytes.NewBuffer(requestJSON),
		// )

		// if reqerr != nil {
		// 	log.Panic("Request creation failed")
		// }
		// req.Header.Set("Content-Type", "application/json")
		// req.Header.Set("Accept", "application/json")

		// // Get the response
		// respPredict, resperr := gohttp.DefaultClient.Do(req)
		// if resperr != nil {
		// 	log.Panic("Request failed")
		// }

		// // Response body converted to stringified JSON
		// respbody, _ := ioutil.ReadAll(respPredict.Body)

		// // Represents the response received from Github
		// type PredictOutput struct {
		// 	Status string `json:"status"`
		// 	Output string `json:"output"`
		// }

		// // Convert stringified JSON to a struct object of type githubAccessTokenResponse
		// var predictResp PredictOutput
		// json.Unmarshal(respbody, &predictResp)

		c.JSON(200, result)
	}
}

func Predict() gin.HandlerFunc {
	return func(c *gin.Context) {
		var resp models.PredictBody

		if err := c.BindJSON(&resp); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var foundModel models.Model
		var foundModel1 models.Model
		err := modelCollection.FindOne(context.TODO(), bson.M{"_id": resp.Id}).Decode(&foundModel)
		if err != nil {
			println(err)
			return
		}

		output, err := json.Marshal(foundModel)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(output, &foundModel1)
		fmt.Println(foundModel1.PredictUrl)
		data := map[string]map[string]string{
			"input": {
				"text": resp.Input,
			},
		}
		requestJSON, _ := json.Marshal(data)

		fmt.Println(string(requestJSON))
		// POST request to set URL
		req, reqerr := gohttp.NewRequest(
			"POST",
			foundModel1.PredictUrl,
			bytes.NewBuffer(requestJSON),
		)

		if reqerr != nil {
			log.Panic("Request creation failed")
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// Get the response
		respPredict, resperr := gohttp.DefaultClient.Do(req)
		if resperr != nil {
			log.Panic("Request failed")
		}

		// Response body converted to stringified JSON
		respbody, _ := ioutil.ReadAll(respPredict.Body)

		// Represents the response received from Github
		type PredictOutput struct {
			Status string `json:"status"`
			Output string `json:"output"`
		}

		// Convert stringified JSON to a struct object of type githubAccessTokenResponse
		var predictResp PredictOutput
		json.Unmarshal(respbody, &predictResp)

		c.JSON(200, predictResp)
	}
}

func HandlerUpload1() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var body models.Message
		defer cancel()
		if err := c.BindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		url := body.Url
		code := body.Code
		claims, _ := helper.DecodeToken(code)
		indexDotCom := strings.LastIndex(url, ".com")
		indexDotGit := strings.LastIndex(url, ".git")
		dir := "repos/" + url[indexDotCom+5:indexDotGit]

		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: "arbruzaz",
				Password: claims.AccessToken,
			},
			URL:               url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "error"})
			return
		}
		dockerImageId, dockerUrl := helper.PushToDocker(dir, body.Name)

		var model models.Model
		model.ID = primitive.NewObjectID()
		model.Name = body.Name
		model.ImageId = dockerImageId
		model.Input = body.Input
		model.Url = dockerUrl
		model.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		helper.CreateDeploy(dockerUrl, body.Name)
		helper.CreateService(body.Name)
		model.PredictUrl, _ = helper.DeployKube(body.Name)
		_, insertErr := modelCollection.InsertOne(ctx, model)
		defer cancel()
		if insertErr != nil {
			c.JSON(500, gin.H{"err": "error"})
		}
		if err != nil {
			c.JSON(500, gin.H{"err": "error"})
		}
		c.JSON(200, gin.H{"message": "Clone Successful"})

	}
}

func GetAllOutputHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("model_id")

		modelID, _ := primitive.ObjectIDFromHex(id)
		var ModelOutputDatas []models.ModelOutputData
		findOptions := options.Find()
		cur, err := modelCollectionOutput.Find(context.TODO(), bson.M{"model_id": modelID}, findOptions)
		if err != nil {
			c.JSON(400, err)
			return
		}

		for cur.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem models.ModelOutputData
			err := cur.Decode(&elem)
			if err != nil {
				c.JSON(400, err)
				return
			}

			ModelOutputDatas = append(ModelOutputDatas, elem)
		}
		if err := cur.Err(); err != nil {
			c.JSON(500, err)
			return
		}

		//Close the cursor once finished
		cur.Close(context.TODO())

		c.JSON(200, ModelOutputDatas)
		return
	}
}

func HandlerUpdateModel() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
func HandlerDeleteModel() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("model_id")

		modelID, _ := primitive.ObjectIDFromHex(id)

		var deletedDocument bson.M
		err := modelCollection.FindOneAndDelete(context.TODO(), bson.M{"model_id": modelID}).Decode(&deletedDocument)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(200, gin.H{"message": "No matched document"})
				return
			}
			c.JSON(500, err)
			return
		}

		err = modelCollectionInputDetail.FindOneAndDelete(context.TODO(), bson.M{"model_id": modelID}).Decode(&deletedDocument)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(200, gin.H{"message": "No matched document"})
				return
			}
			c.JSON(500, err)
			return
		}

		err = modelCollectionImage.FindOneAndDelete(context.TODO(), bson.M{"model_id": modelID}).Decode(&deletedDocument)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(200, gin.H{"message": "No matched document"})
				return
			}
			c.JSON(500, err)
			return
		}
		c.JSON(200, gin.H{"message": "Success"})
	}
}
func HandlerReportModel() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var reportRequest models.ModelReportRequest
		defer cancel()
		if err := c.BindJSON(&reportRequest); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var modelReport models.ModelReport
		modelReport.ModelID, _ = primitive.ObjectIDFromHex(reportRequest.ModelID)
		modelReport.UserID, _ = primitive.ObjectIDFromHex(reportRequest.UserID)
		modelReport.Description = reportRequest.Description
		modelReport.ModelReportID = primitive.NewObjectID()
		modelReport.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		// Insert Input Detail to DB
		defer cancel()
		_, err := modelCollectionReport.InsertOne(ctx, modelReport)
		if err != nil {
			c.JSON(400, err)
			return
		}

		c.JSON(200, gin.H{"message": "Success"})
	}
}

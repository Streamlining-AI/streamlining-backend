package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	gohttp "net/http"
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

func HandlerPredict() {

}

func HandlerUpload1(reqModel models.ModelDataTransfer, githubCode string) models.ModelData {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var model models.ModelData
	model.ModelID = primitive.NewObjectID()
	model.Name = reqModel.Name
	model.Type = reqModel.Type
	model.IsVisible = reqModel.IsVisible
	model.GithubURL = reqModel.GithubURL
	model.Description = reqModel.Description
	model.PredictRecordCount = 0.0
	model.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	model.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	model.UserID, _ = primitive.ObjectIDFromHex(reqModel.UserID)
	model.OutputType = reqModel.Type
	_, err := modelCollection.InsertOne(ctx, model)

	defer cancel()
	if err != nil {
		fmt.Print(err)
	}
	githubCode = reqModel.GithubCode

	claims, _ := helper.DecodeToken(githubCode)
	dir := "repos/" + model.GithubURL

	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: "arbruzaz",
			Password: claims.AccessToken,
		},
		URL:               model.GithubURL,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	ImageID, DockerURL, err := HandlerDeployDocker(dir, model.Name, model.ModelID)
	defer cancel()
	if err != nil {
		fmt.Print(err)
	}
	err = HandlerDeployKube(DockerURL, ImageID, model.Name)
	defer cancel()
	if err != nil {
		fmt.Print(err)
	}

	return model
}

func HandlerDeployDocker(dir string, modelName string, modelID primitive.ObjectID) (primitive.ObjectID, string, error) {
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
		return modelImage.ImageID, DockerURL, err
	}

	return modelImage.ImageID, DockerURL, nil
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

func GetAllModel1() []models.ModelData {
	var foundModels []models.ModelData
	findOptions := options.Find()

	cur, err := modelCollection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		println(err)
		return nil
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
	return foundModels
}

func GetModelByID(modelID string) (models.ModelData, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	paramID := modelID
	var foundModel models.ModelData

	err := modelCollection.FindOne(ctx, bson.M{"model_id": paramID}).Decode(&foundModel)
	defer cancel()
	if err != nil {
		return foundModel, err
	}

	return foundModel, nil
}

func HandlerUpload() gin.HandlerFunc {
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
	}
}

func GetModelById() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		paramID := c.Param("id")
		var foundModel models.Model

		err := modelCollection.FindOne(ctx, bson.M{"_id": paramID}).Decode(&foundModel)
		defer cancel()
		if err != nil {
			c.JSON(500, gin.H{"error": "Model ID is invalid"})
			return
		}

		c.JSON(200, foundModel)

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

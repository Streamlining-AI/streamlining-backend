package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"log"
	gohttp "net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Streamlining-AI/streamlining-backend/database"
	helper "github.com/Streamlining-AI/streamlining-backend/helpers"
	"github.com/Streamlining-AI/streamlining-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/minio/minio-go/v7"
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
		err := os.RemoveAll("repos/")
		if err != nil {
			log.Fatal(err)
		}
		helper.CreateAndGetDir("repos")
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
		model.Registry = reqModel.Registry
		model.PredictRecordCount = 0
		model.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.UserID, _ = primitive.ObjectIDFromHex(reqModel.UserID)
		model.Banner = reqModel.Banner

		// githubCode := reqModel.GithubCode

		// claims, _ := helper.DecodeToken(githubCode)
		dir := "repos/" + model.Name
		descriptDir := "repos/" + model.Name + "/README.md"

		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			// Auth: &http.BasicAuth{
			// 	Username: claims.Username,
			// 	Password: claims.AccessToken,
			// },
			URL:               model.GithubURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		fmt.Println("Git Clone Succesfull")
		if err != nil {
			c.JSON(500, gin.H{"message": "Cannot Get data"})
			return
		}

		description, err := helper.ReadFile(descriptDir)
		if err != nil {
			c.JSON(500, gin.H{"message": "Cannot Get data"})
			return
		}
		model.Description = string(description)

		_, err = modelCollection.InsertOne(ctx, model)
		defer cancel()
		if err != nil {
			fmt.Print(err)
		}
		ImageID, DockerURL, err := HandlerDeployDocker(dir, model.Registry, model.ModelID, reqModel.Model_Version)
		if err != nil {
			c.JSON(500, gin.H{"message": "Error during handling docker image"})
			return
		}

		err = HandlerConfig(dir, model.ModelID, DockerURL)
		if err != nil {
			c.JSON(400, gin.H{"message": "Cannot create config"})
			return
		}

		err = os.RemoveAll("repos/" + model.Name)
		if err != nil {
			log.Fatal(err)
		}
		modelVersion := strings.ReplaceAll(reqModel.Model_Version, ".", "-")
		serviceName := strings.ToLower(model.Name) + modelVersion
		err = HandlerDeployKube(model.Registry, ImageID, strings.ToLower(model.Name), serviceName)

		if err != nil {
			c.JSON(500, gin.H{"message": "Error during handling kubernetes "})
			return
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

	byteResult, _ := io.ReadAll(fileContent)

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
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "output_type", Value: payload.Output.Name}}}}
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

func HandlerDeployDocker(dir string, registry string, modelID primitive.ObjectID, modelVersion string) (primitive.ObjectID, string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	DockerImageID := registry
	fmt.Println("Return from registry" + DockerImageID)
	var modelImage models.ModelImage
	modelImage.ImageID = primitive.NewObjectID()
	modelImage.DockerImageID = DockerImageID
	modelImage.ModelID = modelID
	modelImage.ModelVersion = modelVersion
	modelImage.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	_, err := modelCollectionImage.InsertOne(ctx, modelImage)

	defer cancel()
	if err != nil {
		return modelImage.ImageID, DockerImageID, err
	}

	return modelImage.ImageID, DockerImageID, nil
}

func HandlerDeployKube(DockerURL string, ImageID primitive.ObjectID, modelName string, serviceName string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	err := helper.CreateDeployment(modelName, serviceName, DockerURL)
	if err != nil {
		panic(err)
	}
	fmt.Println("Deployment created successfully.")

	err = helper.CreateService(modelName, serviceName)
	if err != nil {
		panic(err)
	}
	fmt.Println("Service created successfully.")

	PodURL, PredictURL, _ := helper.GetServiceURL(serviceName)
	fmt.Println("Get URL successfully.")

	var modelPod models.ModelPod
	modelPod.PodID = primitive.NewObjectID()
	modelPod.PodURL = PodURL
	modelPod.PredictURL = PredictURL
	modelPod.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	modelPod.ImageID = ImageID

	_, err = modelCollectionPod.InsertOne(ctx, modelPod)

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

		cur, err := modelCollection.Find(context.TODO(), bson.M{"is_visible": true}, findOptions)
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

func GetAllOwnerModel() gin.HandlerFunc {
	return func(c *gin.Context) {

		var foundModels []models.ModelData
		id := c.Param("userid")
		findOptions := options.Find()
		userID, _ := primitive.ObjectIDFromHex(id)
		cur, err := modelCollection.Find(context.TODO(), bson.M{"user_id": userID}, findOptions)
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

func GetModelByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		id := c.Param("model_id")
		docker_image := c.Param("docker_image_id")
		modelID, _ := primitive.ObjectIDFromHex(id)
		var foundModel models.ModelData

		err := modelCollection.FindOne(ctx, bson.M{"model_id": modelID}).Decode(&foundModel)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": "Model ID is invalid"})
			return
		}

		var foundImages []models.ModelImage
		var dockerImagesID []string
		findOptions := options.Find()

		cur, err := modelCollectionImage.Find(context.TODO(), bson.M{"model_id": modelID}, findOptions)
		if err != nil {
			println(err)
			return
		}

		for cur.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem models.ModelImage
			err := cur.Decode(&elem)
			if err != nil {
				log.Fatal(err)
			}

			foundImages = append(foundImages, elem)
		}
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		//Close the cursor once finished
		cur.Close(context.TODO())

		sort.Slice(foundImages, func(i, j int) bool {
			return foundImages[i].CreatedAt.After(foundImages[j].CreatedAt)
		})

		for i := 0; i < len(foundImages); i++ {
			dockerImagesID = append(dockerImagesID, foundImages[i].DockerImageID)
		}

		var modelInput models.ModelInput
		if docker_image == "/" {
			err = modelCollectionInputDetail.FindOne(ctx, bson.M{"model_id": modelID, "docker_image_id": foundImages[0].DockerImageID}).Decode(&modelInput)
		} else {
			err = modelCollectionInputDetail.FindOne(ctx, bson.M{"model_id": modelID, "docker_image_id": docker_image[1:]}).Decode(&modelInput)
		}

		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var modelTransfer models.ModelTransfer

		modelTransfer.ModelID = foundModel.ModelID
		modelTransfer.Name = foundModel.Name
		modelTransfer.Type = foundModel.Type
		modelTransfer.GithubURL = foundModel.GithubURL
		modelTransfer.Description = foundModel.Description
		modelTransfer.PredictRecordCount = foundModel.PredictRecordCount
		modelTransfer.CreatedAt = foundModel.CreatedAt
		modelTransfer.OutputType = foundModel.OutputType
		modelTransfer.DockerImageID = dockerImagesID
		modelTransfer.InputDetail = modelInput.InputDetail

		c.JSON(200, modelTransfer)
	}
}

func GetModelInputByDockerImageID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var reqModel models.ModelIDAndDockerImageID
		defer cancel()
		if err := c.BindJSON(&reqModel); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		modelID, _ := primitive.ObjectIDFromHex(reqModel.ModelID)

		var modelInput models.ModelInput

		err := modelCollectionInputDetail.FindOne(ctx, bson.M{"model_id": modelID, "docker_image_id": reqModel.DockerImageID}).Decode(&modelInput)
		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, modelInput)
	}
}

func HandlerPredict() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var uploadPath = os.TempDir()
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
					// if inputData.DataInputs[i].Type == "image" {
					// 	str := fmt.Sprintf("%v", inputData.DataInputs[i].Data)
					// 	var x interface{} = uploadPath + "/" + str
					// 	// interface{uploadPath + inputData.DataInputs[i].Data}
					// 	inputData.DataInputs[i].Data = x
					// 	str = fmt.Sprintf("%v", inputData.DataInputs[i].Data)

					// }
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

		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "predict_record_count", Value: model.PredictRecordCount + 1}}}}
		_, err = modelCollection.UpdateOne(ctx, bson.M{"model_id": modelID}, update)
		if err != nil {
			fmt.Println(err)
		}
		requestJSON, _ := json.Marshal(predictBody)
		req, reqerr := gohttp.NewRequest(
			"POST",
			modelPod.PredictURL,
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
		respbody, _ := io.ReadAll(respPredict.Body)

		type PredictOutput struct {
			Status string      `json:"status"`
			Output interface{} `json:"output"`
		}

		type PredictOutputWithAttribute struct {
			Status string                 `json:"status"`
			Output map[string]interface{} `json:"output"`
		}

		var predictResp PredictOutput
		var predictRespAtt PredictOutputWithAttribute
		var imgBase64 string
		if model.OutputType == "no" {
			json.Unmarshal(respbody, &predictResp)
			imgBase64 = strings.TrimPrefix(predictResp.Output.(string), "data:image/png;base64,")
		} else {
			json.Unmarshal(respbody, &predictRespAtt)
			imgBase64 = strings.TrimPrefix(predictRespAtt.Output[model.OutputType].(string), "data:image/png;base64,")
		}
		imgBytes, err := base64.StdEncoding.DecodeString(imgBase64)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			log.Fatal(err)
		}
		// Now you can use the decoded image object as needed
		// For example, you can encode it as a PNG and write it to a file

		var fileName = RandToken(12) + ".png"

		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		if err != nil {
			log.Fatal(err)
		}

		var minioClient *minio.Client = database.MinioClient
		// Upload the PNG image to the specified bucket and object
		object := bytes.NewReader(buf.Bytes())
		objectSize := int64(len(buf.Bytes()))

		// Check if the bucket already exists
		exists, err := minioClient.BucketExists(context.Background(), "mybucket")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If the bucket doesn't exist, create it
		if !exists {
			err = minioClient.MakeBucket(context.Background(), "mybucket", minio.MakeBucketOptions{})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Bucket '%s' created successfully.\n", "mybucket")
		} else {
			fmt.Printf("Bucket '%s' already exists.\n", "mybucket")
		}

		_, err = minioClient.PutObject(context.Background(), "mybucket", fileName, object, objectSize, minio.PutObjectOptions{
			ContentType: "image/png",
		})
		if err != nil {
			c.AbortWithStatusJSON(gohttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var modelOutputData models.ModelOutputData
		modelOutputData.ModelOutputID = primitive.NewObjectID()
		modelOutputData.Output = "/files/" + fileName
		modelOutputData.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		modelOutputData.ModelID = modelID
		modelOutputData.ModelInputData = modelInputData

		_, err = modelCollectionOutput.InsertOne(ctx, modelOutputData)
		if err != nil {
			log.Fatal(err)
		}
		// c.JSON(200, gin.H{"output": predictResp})
		c.JSON(200, gin.H{"output": "/files/" + fileName})
	}
}

func HandlerPredictStream() gin.HandlerFunc {
	return func(c *gin.Context) {
		// var uploadPath = os.TempDir()
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
					// if inputData.DataInputs[i].Type == "image" {
					// 	str := fmt.Sprintf("%v", inputData.DataInputs[i].Data)
					// 	var x interface{} = uploadPath + "/" + str
					// 	// interface{uploadPath + inputData.DataInputs[i].Data}
					// 	inputData.DataInputs[i].Data = x
					// 	str = fmt.Sprintf("%v", inputData.DataInputs[i].Data)

					// }
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

		defer cancel()
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		update := bson.D{{Key: "$set", Value: bson.D{{Key: "predict_record_count", Value: model.PredictRecordCount + 1}}}}
		_, err = modelCollection.UpdateOne(ctx, bson.M{"model_id": modelID}, update)
		if err != nil {
			fmt.Println(err)
		}
		requestJSON, _ := json.Marshal(predictBody)
		req, reqerr := gohttp.NewRequest(
			"POST",
			modelPod.PredictURL,
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
		respbody, _ := io.ReadAll(respPredict.Body)

		// Represents the response received from Github
		type PredictOutput struct {
			Status string      `json:"status"`
			Output interface{} `json:"output"`
		}

		type PredictOutputWithAttribute struct {
			Status string                 `json:"status"`
			Output map[string]interface{} `json:"output"`
		}

		var predictResp PredictOutput
		var predictRespAtt PredictOutputWithAttribute
		var imgBase64 string
		if model.OutputType == "no" {
			json.Unmarshal(respbody, &predictResp)
			imgBase64 = strings.TrimPrefix(predictResp.Output.(string), "data:image/png;base64,")
		} else {
			json.Unmarshal(respbody, &predictRespAtt)
			imgBase64 = strings.TrimPrefix(predictRespAtt.Output[model.OutputType].(string), "data:image/png;base64,")
		}

		imgBytes, err := base64.StdEncoding.DecodeString(imgBase64)
		if err != nil {
			log.Fatal(err)
		}
		img, err := png.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			log.Fatal(err)
		}

		var fileName = RandToken(12) + ".png"

		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		if err != nil {
			log.Fatal(err)
		}

		var minioClient *minio.Client = database.MinioClient
		// Upload the PNG image to the specified bucket and object
		object := bytes.NewReader(buf.Bytes())
		objectSize := int64(len(buf.Bytes()))

		// Check if the bucket already exists
		exists, err := minioClient.BucketExists(context.Background(), "mybucket")
		if err != nil {
			fmt.Println(err)
			return
		}

		// If the bucket doesn't exist, create it
		if !exists {
			err = minioClient.MakeBucket(context.Background(), "mybucket", minio.MakeBucketOptions{})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Bucket '%s' created successfully.\n", "mybucket")
		} else {
			fmt.Printf("Bucket '%s' already exists.\n", "mybucket")
		}

		_, err = minioClient.PutObject(context.Background(), "mybucket", fileName, object, objectSize, minio.PutObjectOptions{
			ContentType: "image/png",
		})
		if err != nil {
			c.AbortWithStatusJSON(gohttp.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"output": "/files/" + fileName})
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
	}
}

func HandlerUpdateModel() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := os.RemoveAll("repos/")
		helper.CreateAndGetDir("repos")
		if err != nil {
			log.Fatal(err)
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var reqModel models.ModelUpdateTransfer
		defer cancel()
		if err := c.BindJSON(&reqModel); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var model models.ModelData
		model.ModelID, _ = primitive.ObjectIDFromHex(reqModel.ModelID)
		model.GithubURL = reqModel.GithubURL
		model.Registry = reqModel.Registry
		model.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "updated_at", Value: model.UpdatedAt}}}}
		_, err = modelCollection.UpdateByID(ctx, bson.M{"model_id": model.ModelID}, update)
		defer cancel()
		if err != nil {
			fmt.Print(err)
		}
		// githubCode := reqModel.GithubCode

		// claims, _ := helper.DecodeToken(githubCode)
		dir := "repos/" + reqModel.Name

		_, err = git.PlainClone(dir, false, &git.CloneOptions{
			// Auth: &http.BasicAuth{
			// 	Username: "arbruzaz",
			// 	Password: claims.AccessToken,
			// },
			URL:               model.GithubURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			c.JSON(500, gin.H{"message": "Cannot Get data"})
			return
		}
		ImageID, DockerURL, err := HandlerDeployDocker(dir, model.Registry, model.ModelID, reqModel.Model_Version)
		if err != nil {
			c.JSON(500, gin.H{"message": "Error during handling docker image"})
			return
		}

		err = HandlerConfig(dir, model.ModelID, DockerURL)
		if err != nil {
			c.JSON(400, gin.H{"message": "Cannot create config"})
			return
		}

		err = os.RemoveAll("repos/" + model.Name)
		if err != nil {
			log.Fatal(err)
		}
		modelVersion := strings.ReplaceAll(reqModel.Model_Version, ".", "-")
		serviceName := strings.ToLower(model.Name) + modelVersion
		err = HandlerDeployKube(model.Registry, ImageID, strings.ToLower(model.Name), serviceName)

		if err != nil {
			c.JSON(500, gin.H{"message": "Error during handling kubernetes "})
			return
		}

		c.JSON(200, gin.H{"message": "Clone Successful"})

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

		var foundImages []models.ModelImage
		findOptions := options.Find()

		cur, err := modelCollectionImage.Find(context.TODO(), bson.M{"model_id": modelID}, findOptions)
		if err != nil {
			println(err)
			return
		}

		for cur.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem models.ModelImage
			err := cur.Decode(&elem)
			if err != nil {
				fmt.Print(err)
			}

			foundImages = append(foundImages, elem)
		}
		if err := cur.Err(); err != nil {
			fmt.Print(err)
		}

		//Close the cursor once finished
		cur.Close(context.TODO())
		for i := 0; i < len(foundImages); i++ {
			lastSlashIndex := strings.LastIndex(foundImages[i].DockerImageID, "/")
			if lastSlashIndex >= 0 {
				imageName := strings.Replace(foundImages[i].DockerImageID[lastSlashIndex+1:], ":", "", 1)
				imageName = strings.ReplaceAll(imageName, ".", "-")
				imageName = imageName + "-service"
				err := helper.DeleteDeploymentAndService(imageName, imageName)
				if err != nil {
					fmt.Print(err)
				}
			}

			_, err = modelCollectionPod.DeleteOne(context.TODO(), bson.M{"image_id": foundImages[i].ImageID})
			if err != nil {
				fmt.Print(err)
			}
		}

		_, err = modelCollectionInputDetail.DeleteMany(context.TODO(), bson.M{"model_id": modelID})
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(200, gin.H{"message": "No matched document"})
				return
			}
			c.JSON(500, err)
			return
		}

		_, err = modelCollectionImage.DeleteMany(context.TODO(), bson.M{"model_id": modelID})
		if err != nil {
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

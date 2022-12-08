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
	"os/exec"
	"path/filepath"
	"regexp"
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
	"gopkg.in/yaml.v2"
)

var modelCollection *mongo.Collection = database.OpenCollection(database.Client, "model")

type Message struct {
	Url   string `json:"url"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Input string `json:"input"`
}

func HandlerUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var body Message
		defer cancel()
		if err := c.BindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		url := body.Url
		code := body.Code
		claims, _ := helper.DecodeToken(code)
		// url := "https://github.com/git-fixtures/basic.git"
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
		dockerImageId, dockerUrl := PushToDocker(dir, body.Name)

		var model models.Model
		model.ID = primitive.NewObjectID()
		model.Name = body.Name
		model.ImageId = dockerImageId
		model.Input = body.Input
		model.Url = dockerUrl
		model.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		CreateDeploy(dockerUrl, body.Name)
		CreateService(body.Name)
		model.PredictUrl = DeployKube(body.Name)
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

func PushToDocker(folderName string, name string) (string, string) {
	libRegEx, e := regexp.Compile("cog.yaml")
	if e != nil {
		log.Fatal(e)
	}

	e = filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err == nil && libRegEx.MatchString(info.Name()) {
			println(info.Name(), path)
			// pathModel = path
			println(folderName)
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}

	commandCog := "cog"
	task := "build"
	cogTag := "-t"

	cmd := exec.Command(commandCog, task, cogTag, name)
	cmd.Dir = folderName
	if err := cmd.Run(); err != nil {
		fmt.Println("could not run command: ", err)
	}

	commandDocker := "docker"
	images := "images"
	login := "login"
	tag := "tag"
	push := "push"
	ip := "192.168.49.2:30001"
	userAuth := "-u"
	user := "admin"
	passAuth := "-p"
	password := "password"
	// docker images name
	cmd1, err := exec.Command(commandDocker, images, name).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput := string(cmd1)
	dockerDescript := strings.Split(dockerOutput, "\n")
	dockerDetail := strings.Fields(dockerDescript[1])
	dockerImageId := dockerDetail[2]

	fmt.Println(dockerImageId)

	// docker login 192.168.49.2:30001 -u 'admin' -p 'password'
	cmd2, err := exec.Command(commandDocker, login, ip, userAuth, user, passAuth, password).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd2)
	fmt.Println(dockerOutput)

	// docker tag {ImageID} 192.168.49.2:30001/{imageName}
	cmd3, err := exec.Command(commandDocker, tag, dockerImageId, ip+"/"+dockerImageId).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd3)
	fmt.Println(dockerOutput)

	// docker push 192.168.49.2:30001/{imageName}
	cmd4, err := exec.Command(commandDocker, push, ip+"/"+dockerImageId).Output()
	if err != nil {
		log.Fatal(err)
	}
	dockerOutput = string(cmd4)
	fmt.Println(dockerOutput)
	return dockerImageId, ip + "/" + dockerImageId
}

func CreateDeploy(URL string, Name string) {
	type MetadataStruct struct {
		Name string `yaml:"name"`
	}

	type MatchLabelsStruct struct {
		App string `yaml:"app"`
	}

	// ==============================================
	type SelectorStruct struct {
		MatchLabels MatchLabelsStruct `yaml:"matchLabels"`
	}

	type LabelsStruct struct {
		App string `yaml:"app"`
	}

	type TemplateMetadataStruct struct {
		Labels LabelsStruct `yaml:"labels"`
	}

	// ==============================================

	type PortsStruct struct {
		ContainerPort int `yaml:"containerPort"`
	}

	type ContainersStruct struct {
		Name  string `yaml:"name"`
		Image string `yaml:"image"`
		Ports []PortsStruct
	}

	// ==============================================

	type ImagePullSecretsStruct struct {
		Name string `yaml:"name"`
	}

	type TemplateSpecStruct struct {
		Containers       []ContainersStruct       `yaml:"containers"`
		ImagePullSecrets []ImagePullSecretsStruct `yaml:"imagePullSecrets"`
	}

	type TemplateStruct struct {
		Metadata TemplateMetadataStruct `yaml:"metadata"`
		Spec     TemplateSpecStruct     `yaml:"spec"`
	}

	type SpecStruct struct {
		Replicas int            `yaml:"replicas"`
		Selector SelectorStruct `yaml:"selector"`
		Template TemplateStruct `yaml:"template"`
	}

	type Deploy struct {
		Kind       string         `yaml:"kind"`
		ApiVersion string         `yaml:"apiVersion"`
		Metadata   MetadataStruct `yaml:"metadata"`
		Spec       SpecStruct     `yaml:"spec"`
	}
	s1 := Deploy{
		Kind:       "Deployment",
		ApiVersion: "apps/v1",
		Metadata: MetadataStruct{
			Name: Name,
		},
		Spec: SpecStruct{
			Replicas: 1,
			Selector: SelectorStruct{
				MatchLabels: MatchLabelsStruct{
					App: Name,
				},
			},
			Template: TemplateStruct{
				Metadata: TemplateMetadataStruct{
					Labels: LabelsStruct{
						App: Name,
					},
				},
				Spec: TemplateSpecStruct{
					Containers: []ContainersStruct{{
						Name:  Name,
						Image: URL,
						Ports: []PortsStruct{{
							ContainerPort: 5000,
						}},
					}},
					ImagePullSecrets: []ImagePullSecretsStruct{
						{
							Name: "regcred",
						},
					},
				},
			},
		},
	}

	yamlData1, err := yaml.Marshal(&s1)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	err2 := ioutil.WriteFile("Deployment.yaml", yamlData1, 0777)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func CreateService(Name string) {
	type LabelsStruct struct {
		App string `yaml:"app"`
	}

	type MetadataStruct struct {
		Name   string       `yaml:"name"`
		Labels LabelsStruct `yaml:"labels"`
	}

	type PortsStruct struct {
		Port       int    `yaml:"port"`
		TargetPort int    `yaml:"targetPort"`
		Protocol   string `yaml:"protocol"`
	}

	type SelectorStruct struct {
		App string `yaml:"app"`
	}

	type SpecStruct struct {
		Type     string `yaml:"type"`
		Ports    []PortsStruct
		Selector SelectorStruct `yaml:"selector"`
	}

	type Service struct {
		ApiVersion string         `yaml:"apiVersion"`
		Kind       string         `yaml:"kind"`
		Metadata   MetadataStruct `yaml:"metadata"`
		Spec       SpecStruct     `yaml:"spec"`
	}

	s2 := Service{
		ApiVersion: "v1",
		Kind:       "Service",
		Metadata: MetadataStruct{
			Name: Name + "-service",
			Labels: LabelsStruct{
				App: Name,
			},
		},
		Spec: SpecStruct{
			Type: "LoadBalancer",
			Ports: []PortsStruct{
				{
					Port:       5000,
					TargetPort: 5000,
					Protocol:   "TCP",
				},
			},
			Selector: SelectorStruct{
				App: Name,
			},
		},
	}

	yamlData1, err := yaml.Marshal(&s2)

	if err != nil {
		fmt.Printf("Error while Marshaling. %v", err)
	}
	err2 := ioutil.WriteFile("Service.yaml", yamlData1, 0777)

	if err2 != nil {

		log.Fatal(err2)
	}
}

func DeployKube(Name string) string {
	cmd5, err := exec.Command("kubectl", "apply", "-f", ".").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cmd5)

	cmd6, err := exec.Command("minikube", "service", "--url", Name+"-service").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cmd6)

	return string(cmd6) + "/predictions"
}

func GetAllModel() gin.HandlerFunc {
	return func(c *gin.Context) {
		var foundModels []models.Model
		findOptions := options.Find()

		cur, err := modelCollection.Find(context.TODO(), bson.D{{}}, findOptions)
		if err != nil {
			println(err)
			return
		}

		for cur.Next(context.TODO()) {
			//Create a value into which the single document can be decoded
			var elem models.Model
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

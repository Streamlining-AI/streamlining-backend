package controllers

import (
	"context"
	"fmt"
	"log"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
		dockerImageId, dockerUrl := PushToDocker(dir,body.Name)

		var model models.Model
		model.ID = primitive.NewObjectID()
		model.Name = body.Name
		model.ImageId = dockerImageId
		model.Input = body.Input
		model.Url = dockerUrl
		model.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		model.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
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

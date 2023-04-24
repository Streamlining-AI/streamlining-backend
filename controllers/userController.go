package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/Streamlining-AI/streamlining-backend/database"

	helper "github.com/Streamlining-AI/streamlining-backend/helpers"
	"github.com/Streamlining-AI/streamlining-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollectionGithub *mongo.Collection = database.OpenCollection(database.Client, "userGithub")
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "login or passowrd is incorrect"
		check = false
	}

	return check, msg
}

// CreateUser is the api used to tget a single user
func Regsiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User

		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		defer cancel()
		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		defer cancel()
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or passowrd is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, foundUser.User_id)

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("token", token, 3600, "/", "localhost:3000", true, true)

		c.JSON(http.StatusOK, foundUser)

	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("token", "", -1, "/", "103.153.118.69:30003", true, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})

	}
}

func RegisterGithub(userName string, userId string) (string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var user models.UserGithub
	var err error
	user.ID = primitive.NewObjectID()
	user.User_id = userId
	user.Username = userName
	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	defer cancel()
	if err != nil {
		return "", err
	}

	_, insertErr := userCollectionGithub.InsertOne(ctx, user)
	defer cancel()
	if insertErr != nil {
		return "", insertErr
	}

	return user.ID.String(), nil
}

func GithubLoginHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		githubClientID := helper.GetGithubClientID()

		redirectURL := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo,user",
			githubClientID,
			"http://103.153.118.69:30003/login/github/callback",
		)
		c.JSON(http.StatusOK, gin.H{"redirectURL": redirectURL})
	}
}

func GithubCallbackHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		var resp models.GithubRequestBody
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()
		if err := c.BindJSON(&resp); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		githubAccessToken := helper.GetGithubAccessToken(resp.Code)
		githubData := GetGithubData(githubAccessToken)

		var prettyJSON bytes.Buffer
		parserr := json.Indent(&prettyJSON, []byte(githubData), "", "\t")
		if parserr != nil {
			log.Panic("JSON parse error")
		}

		if strings.Contains(string(prettyJSON.String()), `"message": "Bad credentials"`) {
			println("Bad Credentials")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Bad Credentials"})
			return
		}
		var data map[string]interface{}

		// Unmarshal the JSON string into the data map
		err := json.Unmarshal([]byte(githubData), &data)
		if err != nil {
			fmt.Println(err)
			return
		}

		username := data["login"]
		githubID := data["id"]
		githubID = int(githubID.(float64))

		strUsername := fmt.Sprintf("%v", username)
		strGithubID := fmt.Sprintf("%d", githubID)

		token, err := helper.EncodeToken(githubAccessToken, strGithubID, strUsername)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		count, err := userCollectionGithub.CountDocuments(ctx, bson.M{"username": strUsername})

		defer cancel()
		if err != nil {
			msg := "error occured while checking for the username"
			c.JSON(http.StatusBadRequest, msg)
			return
		}
		var userID string
		var user models.UserGithub
		if count == 0 {

			userID, err = RegisterGithub(strUsername, strGithubID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}

		} else {
			err := userCollectionGithub.FindOne(ctx, bson.M{"username": strUsername}).Decode(&user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			userID = user.ID.String()
		}

		// ====================================================================================
		/*
			userData := strings.Split(githubData, ",")
			userName := strings.Split(userData[0], ":")
			userName1 := strings.Trim(userName[1], `"`)

			userId := strings.Split(userData[1], ":")
			userId1 := userId[1]

			count, err := userCollectionGithub.CountDocuments(ctx, bson.M{"username": userName1})

			defer cancel()
			if err != nil {
				msg := "error occured while checking for the username"
				c.JSON(http.StatusBadRequest, msg)
				return
			}

			if count == 0 {

				err = RegisterGithub(userName1, userId1)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}

			}
			token, err := helper.EncodeToken(githubAccessToken, userId1)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		*/
		// =================================================================
		c.SetSameSite(http.SameSiteNoneMode)
		c.SetCookie("token", token, 1000*60*60*24, "/", "103.153.118.69:30003", true, true)
		c.JSON(http.StatusOK, gin.H{"token": token, "ID": userID, "username": strUsername})
	}
}

func GetGithubData(accessToken string) string {
	// Get request to a set URL
	req, reqerr := http.NewRequest(
		"GET",
		"https://api.github.com/user",
		nil,
	)
	if reqerr != nil {
		log.Panic("API Request creation failed")
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	// Read the response as a byte slice
	respbody, _ := io.ReadAll(resp.Body)

	// Convert byte slice to string and return
	// util.GitClone(accessToken)
	return string(respbody)
}

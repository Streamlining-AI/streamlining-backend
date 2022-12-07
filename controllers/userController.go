package controllers

import (
	"context"
	"fmt"
	"log"

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
		msg = fmt.Sprintf("login or passowrd is incorrect")
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

		user.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		defer cancel()
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
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
		if passwordIsValid != true {
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

		c.SetCookie("token", "", -1, "/", "localhost:3000", false, true)
		c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})

	}
}

func RegisterGithub(userName string, userId string) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var user models.UserGithub
	var err error
	user.ID = primitive.NewObjectID()
	user.User_id = userId
	user.Username = userName
	user.Created_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	user.Updated_at, err = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	defer cancel()
	if err != nil {
		return err
	}

	_, insertErr := userCollectionGithub.InsertOne(ctx, user)
	defer cancel()
	if insertErr != nil {
		return insertErr
	}

	return nil
}

func GithubLoginHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		githubClientID := helper.GetGithubClientID()

		redirectURL := fmt.Sprintf(
			"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo,user",
			githubClientID,
			"https://localhost:3000/login/github/callback",
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

		count, err := userCollectionGithub.CountDocuments(ctx, bson.M{"username": resp.Username})

		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("error occured while checking for the username")
			c.JSON(http.StatusBadRequest, msg)
			return
		}

		if count == 0 {
			err = RegisterGithub(resp.Username, resp.UserId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		}

		// githubAccessToken := helper.GetGithubAccessToken(resp.Code)
		githubAccessToken := "12345"
		token, err := helper.EncodeToken(githubAccessToken, resp.UserId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		// claims, _ := helper.DecodeToken(token)
		c.SetCookie("token", token, 3600, "/", "127.0.0.1", false, true)
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

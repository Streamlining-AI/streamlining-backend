package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"net/http"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
)

func GetGithubClientID() string {

	githubClientID, exists := os.LookupEnv("CLIENT_ID")
	if !exists {
		log.Fatal("Github Client ID not defined in .env file")
	}

	return githubClientID
}

func GetGithubClientSecret() string {

	githubClientSecret, exists := os.LookupEnv("CLIENT_SECRET")
	if !exists {
		log.Fatal("Github Client ID not defined in .env file")
	}

	return githubClientSecret
}

func GetGithubAccessToken(code string) string {

	clientID := GetGithubClientID()
	clientSecret := GetGithubClientSecret()

	// Set us the request body as JSON
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	// POST request to set URL
	req, reqerr := http.NewRequest(
		"POST",
		"https://github.com/login/oauth/access_token",
		bytes.NewBuffer(requestJSON),
	)
	if reqerr != nil {
		log.Panic("Request creation failed")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Get the response
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		log.Panic("Request failed")
	}

	// Response body converted to stringified JSON
	respbody, _ := ioutil.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return ghresp.AccessToken
}

type TokenDetail struct {
	AccessToken string
	Uid         string
	Username    string
	jwt.StandardClaims
}

func EncodeToken(accessToken string, uid string, username string) (signedToken string, err error) {
	SECRET_KEY, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Fatal("SECRET_KEY not defined in .env file")
	}

	claims := &TokenDetail{
		AccessToken: accessToken,
		Uid:         uid,
		Username:    username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, err
}

func DecodeToken(accessToken string) (claims *TokenDetail, msg string) {
	SECRET_KEY, exists := os.LookupEnv("SECRET_KEY")
	if !exists {
		log.Fatal("SECRET_KEY not defined in .env file")
	}

	token, err := jwt.ParseWithClaims(
		accessToken,
		&TokenDetail{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		msg = err.Error()

	}

	claims, ok := token.Claims.(*TokenDetail)
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")

	}
	return claims, msg
}

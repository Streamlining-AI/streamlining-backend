package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type Message struct {
	Url  string `json:"url"`
	Code string `json:"code"`
}

func HandlerUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body Message

		if err := c.BindJSON(&body); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		url := body.Url
		code := body.Code
		// url := "https://github.com/git-fixtures/basic.git"
		indexDotCom := strings.LastIndex(url, ".com")
		indexDotGit := strings.LastIndex(url, ".git")
		dir := url[indexDotCom+5 : indexDotGit]

		_, err := git.PlainClone(dir, false, &git.CloneOptions{
			Auth: &http.BasicAuth{
				Username: "arbruzaz",
				Password: code,
			},
			URL:               url,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err != nil {
			c.JSON(500, gin.H{"error": "error"})
			return
		}
		c.JSON(200, gin.H{"message": "Clone Successful"})
	}
}

package routes

import (
	controller "github.com/Streamlining-AI/streamlining-backend/controllers"

	"github.com/gin-gonic/gin"
)

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/register", controller.Regsiter())
	incomingRoutes.POST("/users/login", controller.Login())
	incomingRoutes.GET("/users/login/github", controller.GithubLoginHandler())
	incomingRoutes.GET("/users/login/github/callback", controller.GithubCallbackHandler())
}

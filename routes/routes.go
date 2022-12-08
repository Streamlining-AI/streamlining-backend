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
	incomingRoutes.POST("/users/login/github/callback", controller.GithubCallbackHandler())
	incomingRoutes.GET("/model/:id", controller.GetModelById())
	incomingRoutes.GET("/model/", controller.GetAllModel())
}

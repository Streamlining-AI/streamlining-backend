package routes

import (
	controller "github.com/Streamlining-AI/streamlining-backend/controllers"
	helper "github.com/Streamlining-AI/streamlining-backend/helpers"
	ModelSchema "github.com/Streamlining-AI/streamlining-backend/schemas/model"
	"github.com/gin-gonic/gin"
)

// UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/users/register", controller.Regsiter())
	incomingRoutes.POST("/users/login", controller.Login())
	incomingRoutes.GET("/users/login/github", controller.GithubLoginHandler())
	incomingRoutes.POST("/users/login/github/callback", controller.GithubCallbackHandler())
	// incomingRoutes.GET("/model/:id", controller.GetModelById())
	// incomingRoutes.GET("/model/", controller.GetAllModel())
	incomingRoutes.POST("/predict/", controller.Predict())
}

func ModelRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/model", func(c *gin.Context) {
		result := helper.ExecuteQuery(c.Query("query"), ModelSchema.Schema)
		c.JSON(200, result)
	})
}

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
	incomingRoutes.GET("/model/:id", controller.GetModelByID())
	incomingRoutes.GET("/model/", controller.GetAllModel())
	incomingRoutes.POST("/predict/", controller.HandlerPredict())
	incomingRoutes.GET("/model/output/:model_id", controller.GetAllOutputHistory())
	incomingRoutes.DELETE("/model/:model_id", controller.HandlerDeleteModel())
	incomingRoutes.POST("/model/", controller.HandlerUpload())
	incomingRoutes.POST("/model/report", controller.HandlerReportModel())
}

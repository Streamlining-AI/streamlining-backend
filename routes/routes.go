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
}

func ModelRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/model/:model_id/*docker_image_id", controller.GetModelByID())
	incomingRoutes.GET("/model/input", controller.GetModelInputByDockerImageID())
	incomingRoutes.GET("/model/", controller.GetAllModel())
	incomingRoutes.GET("/users/model/:userid", controller.GetAllOwnerModel())
	incomingRoutes.POST("/predict/", controller.HandlerPredict())
	incomingRoutes.POST("/stream/", controller.HandlerPredictStream())
	incomingRoutes.GET("/model/output/:model_id", controller.GetAllOutputHistory())
	incomingRoutes.DELETE("/model/:model_id", controller.HandlerDeleteModel())
	incomingRoutes.POST("/model/", controller.HandlerUpload())
	incomingRoutes.POST("/model/report", controller.HandlerReportModel())
	incomingRoutes.PUT("/model/", controller.HandlerUpdateModel())
}

func ImageRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("/upload", controller.UploadFileHandler())
	incomingRoutes.GET("/files/:filename", controller.GetFile())
}

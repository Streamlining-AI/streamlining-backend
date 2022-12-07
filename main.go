package main

import (
	"log"

	controller "github.com/Streamlining-AI/streamlining-backend/controllers"
	// middleware "github.com/Streamlining-AI/streamlining-backend/middleware"
	routes "github.com/Streamlining-AI/streamlining-backend/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/joho/godotenv"
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}
}

func main() {
	router := gin.New()
	router.Use(gin.Logger())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "OPTION"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.UserRoutes(router)

	router.GET("/upload/clone", controller.HandlerUpload())
	// router.Use(middleware.Authentication())

	// API-2

	// API-1
	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.GET("/users/logout", controller.Logout())

	router.Run(":" + "8000")
}

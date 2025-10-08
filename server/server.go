package server

import (
	"log"
	"xstudious-guide/ai"
	"xstudious-guide/amazon"
	"xstudious-guide/email"
	location "xstudious-guide/maps"

	"github.com/gin-gonic/gin"
)

func InitServer() {
	go hub.Run()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// connect DynamoDB
	dynamoClient, dynamodbStatus := amazon.ConnectDB()
	AddDynamoDBRoutes(dynamoClient, router)

	// connect S3
	s3Client, s3Status := amazon.ConnectS3()
	AddS3Routes(s3Client, dynamoClient, router)

	// connect Google Maps
	mapClient, mapsStatus := location.InitMaps()
	AddMapRoutes(mapClient, router)

	// connect with OpenAI
	aiClient, openAIStatus := ai.InitAi()
	AddAIROutes(aiClient, router)

	// commect with Resend
	emailClient, emailStatus := email.InitEmail()
	AddEmailRoutes(emailClient, router)

	router.GET("/ws", serveWs)
	router.POST("/webhook", WebhookHandler)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"DynamoDB":    dynamodbStatus,
			"S3":          s3Status,
			"Google_Maps": mapsStatus,
			"OpenAI":      openAIStatus,
			"Resend":      emailStatus,
		})
	})

	// Start the server
	log.Println("Server listening on :8080")
	router.Run(":8080")
}

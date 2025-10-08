package server

import (
	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
)

func SendEmailHandler(client *resend.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var email resend.SendEmailRequest
		if err := c.ShouldBindJSON(&email); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		_, err := client.Emails.Send(&email)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to send email"})
			return
		}

		c.JSON(200, gin.H{"message": "Email sent successfully"})
	}
}

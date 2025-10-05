package server

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WebhookHandler(c *gin.Context) {

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}

	// Process the webhook payload as needed

	c.JSON(http.StatusOK, gin.H{"status": "success", "received": string(body)})
}

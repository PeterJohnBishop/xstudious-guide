package server

import (
	"context"
	"net/http"
	"strings"
	"xstudious-guide/authentication"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
)

// gpt-5	OpenAI’s newest general-purpose model. Highest quality / understanding. Likely higher cost.
// gpt-5-mini	A lighter / cheaper variant of gpt-5. Good if you want much of the power but lower cost / latency.
// gpt-4.1	The next iteration of GPT-4-series. Improved long-context handling, coding / instruction following.
// gpt-4.1-mini / gpt-4.1-nano	Smaller / more efficient versions of GPT-4.1. Trade off a bit of capability for cost / speed.
// gpt-4o	Multimodal model (can handle text + images + audio). Useful if you’ll at some point feed images/audio or want richer interaction.
// gpt-4o-mini	A cheaper faster version of gpt-4o, with reduced cost / fewer capabilities.
// o3, o3-mini	Reasoning-focused models. Good for logic, math, etc. Less “multimodal” stuff, more specialized for text reasoning.
// o4-mini	A newer reasoning-model variant (lighter / cheaper) that improves over earlier “o”-series for certain tasks.

type BasicPrompt struct {
	Prompt string `json:"prompt"`
}

func SendPrompt(client *openai.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var prompt BasicPrompt
		if err := c.ShouldBindJSON(&prompt); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if c.Request.Method != http.MethodPost {
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
			return
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}
		claims := authentication.ParseAccessToken(token)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to verify token"})
			return
		}

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-5",
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt.Prompt,
					},
				},
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create access token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Prompt Sent!",
			"response": resp.Choices[0].Message.Content,
		})

	}
}

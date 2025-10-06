package ai

import (
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func InitAi() *openai.Client {

	key := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(key)
	log.Printf("Connected to OpenAI\n")
	return client
}

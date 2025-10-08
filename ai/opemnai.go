package ai

import (
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func InitAi() (*openai.Client, string) {

	key := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(key)
	if client == nil {
		msg := "unable to connect to OpenAI"
		return nil, msg
	}
	return client, "Connected to OpenAI"
}

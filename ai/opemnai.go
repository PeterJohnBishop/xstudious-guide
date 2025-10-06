package ai

import (
	"log"
	"os"

	"github.com/openai/openai-go" // imported as openai
	"github.com/openai/openai-go/option"
)

func InitAi() *openai.Client {

	key := os.Getenv("OPENAI_API_KEY")

	client := openai.NewClient(
		option.WithAPIKey(key),
	)
	log.Printf("Connected to OpenAI\n")
	return &client
}

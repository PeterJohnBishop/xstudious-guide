package email

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v2"
)

func SendEmail() error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Peter <me@peterjohnbishop.com>",
		To:      []string{"peterjbishop.denver@gmail.com"},
		Subject: "Hello world",
		Html:    "<strong>It works!</strong>",
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
	return nil
}

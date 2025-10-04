package email

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEmail(alias string, email string, recipients []string, subject string, html string) error {

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	client := resend.NewClient(apiKey)

	sender := fmt.Sprintf("%s <%s>", alias, email)

	params := &resend.SendEmailRequest{
		From:    sender,
		To:      recipients,
		Subject: subject,
		Html:    html,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
	return nil
}

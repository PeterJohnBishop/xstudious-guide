package email

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailRequest struct {
	Alias      string   `json:"alias"`
	Sender     string   `json:"sender"`
	Recipients []string `json:"recipients"`
	Subject    string   `json:"subject"`
	Html       string   `json:"html"`
}

func InitEmail() (*resend.Client, string) {

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		msg := "RESEND_API_KEY environment variable is not set"
		return nil, msg
	}

	client := resend.NewClient(apiKey)
	if client == nil {
		msg := "unable to connect to Resend"
		return nil, msg
	}
	return client, "Connected to Resend"
}

func SendEmail(client *resend.Client, email EmailRequest) error {

	sender := fmt.Sprintf("%s <%s>", email.Alias, email.Sender)

	params := &resend.SendEmailRequest{
		From:    sender,
		To:      email.Recipients,
		Subject: email.Subject,
		Html:    email.Html,
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(sent.Id)
	return nil
}

// alias := "Acme"
// sender := "go-server@peterjohnbishop.com"
// recipients := []string{"peterjbishop.denver@gmail.com"}
// subject := "Hello world"
// html := "<strong>It works!</strong>"

// err = email.SendEmail(alias, sender, recipients, subject, html)
// if err != nil {
// 	fmt.Println("Error sending email:", err)
// } else {
// 	fmt.Println("Email sent successfully")
// }

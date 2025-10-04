package main

import (
	"fmt"
	"xstudious-guide/email"
)

func main() {

	alias := "Acme"
	sender := "go-server@peterjohnbishop.com"
	recipients := []string{"peterjbishop.denver@gmail.com"}
	subject := "Hello world"
	html := "<strong>It works!</strong>"

	err := email.SendEmail(alias, sender, recipients, subject, html)
	if err != nil {
		fmt.Println("Error sending email:", err)
	} else {
		fmt.Println("Email sent successfully")
	}
}

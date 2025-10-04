package main

import (
	"fmt"
	"log"
	"xstudious-guide/amazon"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg := amazon.StartAws()
	s3Client := amazon.ConnectS3(cfg)
	fmt.Println(s3Client)
	dynamoClient := amazon.ConnectDB(cfg)
	fmt.Println(dynamoClient)

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
}

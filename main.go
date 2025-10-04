package main

import (
	"fmt"
	"xstudious-guide/email"
)

func main() {
	err := email.SendEmail()
	if err != nil {
		fmt.Println("Error sending email:", err)
	} else {
		fmt.Println("Email sent successfully")
	}
}

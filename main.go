package main

import (
	"log"
	"xstudious-guide/server"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, proceeding with environment variables")
	}

	server.InitServer()

}

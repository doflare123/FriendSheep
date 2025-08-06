package main

import (
	"log"
	"notify_service/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables from container")
	}
}

func main() {
	if err := db.InitDatabase(); err != nil {
		log.Fatal(err)
	}
	if err := db.InitFCM(); err != nil {
		log.Fatal(err)
	}
	r := gin.Default()

	r.Run(":8083")
}

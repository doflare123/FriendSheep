package main

import (
	"log"
	"notify_service/db"
	worker "notify_service/schedulers"
	"os"

	"github.com/gin-gonic/gin"
)


func main() {
	r := gin.Default()

	db.InitDatabase()
	db.SeedNotificationTypes(db.GetDB())
	worker.StartNotificationWorker(db.GetDB())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

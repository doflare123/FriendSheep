package main

import (
	"friendship/db"
	"friendship/routes"
	"friendship/services"
	"friendship/utils"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}
}

func main() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("password", utils.PasswordValidation)
		services.InitValidator(v)
	}
	// Создаем новый экземпляр роутера
	r := gin.Default()
	db.InitDatabase()
	db.InitMongoDB()
	routes.RoutesUsers(r)
	// Запускаем сервер на порту 8080
	r.Run(":8080")
}

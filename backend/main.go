package main

import (
	"friendship/db"
	"friendship/routes"
	"friendship/services"
	"friendship/utils"
	"log"
	"net/http"

	_ "friendship/docs"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}
}

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer <your_token>
func main() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("password", utils.PasswordValidation)
		services.InitValidator(v)
	}
	// Создаем новый экземпляр роутера
	r := gin.Default()
	db.InitDatabase()
	db.SeedCategories()
	db.InitMongoDB()
	routes.RoutesUsers(r)
	routes.RoutesAuth(r)
	routes.RouterGroups(r)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	// Запускаем сервер на порту 8080
	r.Run(":8080")
}

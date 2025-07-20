package main

import (
	"friendship/db"
	_ "friendship/docs"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services"
	"friendship/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
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
		_ = v.RegisterValidation("username", utils.ValidateNameTag)
		services.InitValidator(v)
	}

	utils.Init()
	r := gin.New()

	r.Use(middlewares.ErrorLogger())

	// Вариант 1: Простое логирование
	// r.Use(middlewares.RequestResponseLogger())

	// Вариант 2: Детальное логирование
	r.Use(middlewares.DetailedRequestLogger())

	// Вариант 3: Логирование медленных запросов
	//r.Use(middlewares.SlowRequestLogger(2 * time.Second))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.OPTIONS("/*path", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})

	r.Static("/uploads", "./uploads")

	// Инициализация БД и маршрутов
	db.InitDatabase()
	db.SeedCategories()
	db.SeedCategoriesSessionsVisibility()
	db.InitMongoDB()

	routes.RoutesUsers(r)
	routes.RoutesAuth(r)
	routes.RouterGroups(r)
	routes.RouterSessions(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Логируем запуск сервера
	utils.Log.Println("INFO: Starting server on :8080")
	r.Run(":8080")
}

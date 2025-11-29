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
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


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
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001", "https://friendsheep.ru"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
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
	db.SeedStatusSessions()
	db.SeedGenres()
	db.SeedDays()
	db.InitMongoDB()
	if err := db.InitRedis(); err != nil {
		log.Fatal("Ошибка инициализации Redis:", err)
	}

	if err := services.InitPopularSessionsCache(); err != nil {
		log.Fatal("Ошибка инициализации кэша популярных сессий:", err)
	}

	defer func() {
		services.StopPopularSessionsCache()
	}()
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")

	if err := middlewares.InitS3(s3AccessKey, s3SecretKey); err != nil {
		utils.Log.Println("Не удалось инициализировать S3: ", err)
	}
	utils.Log.Println("S3 клиент успешно инициализирован")

	routes.RoutesUsers(r)
	routes.RoutesAuth(r)
	routes.RouterGroups(r)
	routes.RouterSessions(r)
	routes.RouterUsersInfo(r)
	routes.RoutesNews(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Логируем запуск сервера
	utils.Log.Println("INFO: Starting server on :8080")
	r.Run(":8080")
}

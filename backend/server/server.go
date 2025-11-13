package server

import (
	"friendship/config"
	"friendship/db"
	"friendship/logger"
	"friendship/models"
	"friendship/repository"
	session "friendship/sessions"
	"log"
	"net/http"

	_ "friendship/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine       *gin.Engine
	logger       logger.Logger
	postgres     repository.PostgresRepository
	mongo        repository.MongoRepository
	sessionStore session.SessionStore
	cfg          config.Config
}

func InitServer() (*Server, error) {
	conf := config.NewConfig()
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	logger.Info("Logger and config initialized")
	logger.Debug("Что лежит в мыле: ", "email", conf.Email.From, "password", conf.Email.Password, "host", conf.Email.SmtpHost, "port", conf.Email.SmtpPort)
	postgres := repository.NewSheepRepository(logger, conf)
	mongo := repository.NewMongoRepository(logger, conf)
	redis := repository.NewRedisRepository(logger, conf)
	sessionStore := session.NewSessionStore(redis)
	if conf.AppEnv == "DEV" {
		gin.SetMode(gin.DebugMode)
		if err := db.AutoMigDB(postgres, &models.User{}); err != nil {
			logger.Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := db.MigrationDB(postgres, logger); err != nil {
			logger.Error("Error with migrations: %s", err)
		}
	}
	r := gin.Default()
	// r.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:3000"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Access-Control-Allow-Origin"},
	// 	ExposeHeaders:    []string{"Content-Length"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "../swagger/index.html")
	})

	s := &Server{
		engine:       r,
		logger:       logger,
		postgres:     postgres,
		mongo:        mongo,
		sessionStore: sessionStore,
		cfg:          *conf,
	}
	s.initRouters()
	logger.Info("Server init")
	return s, nil
}

func (s *Server) Run(addr string) error {
	s.logger.Info("Server running on " + addr)
	return s.engine.Run(addr)
}

package server

import (
	"errors"
	storage "friendship/S3"
	"friendship/config"
	"friendship/db"
	"friendship/logger"
	"friendship/repository"
	event "friendship/services/events"
	session "friendship/sessions"
	"friendship/validator"
	"log"
	"net/http"

	_ "friendship/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine               *gin.Engine
	logger               logger.Logger
	postgres             repository.PostgresRepository
	S3                   storage.S3Storage
	mongo                repository.MongoRepository
	sessionStore         session.SessionStore
	validators           *validator.Validator
	cfg                  config.Config
	popularEventsService event.PopularEventsService
}

func InitServer() (*Server, error) {
	conf := config.NewConfig()
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	logger.Info("Logger and config initialized")
	postgres := repository.NewSheepRepository(logger, conf)
	mongo := repository.NewMongoRepository(logger, conf)
	redis := repository.NewRedisRepository(logger, conf)
	sessionStore := session.NewSessionStore(redis)
	hasSQLMigrations, err := db.HasMigrationSource()
	if err != nil {
		logger.Error("Error checking migration source", "error", err)
		return nil, err
	}
	if conf.AppEnv == "DEV" {
		gin.SetMode(gin.DebugMode)
		if err := db.BootstrapRegistrationSchema(postgres); err != nil {
			logger.Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if hasSQLMigrations {
			if err := db.MigrationDB(postgres, logger); err != nil {
				logger.Error("Error with migrations: %s", err)
				return nil, err
			}
			hasCoreSchema, err := db.HasCoreSchemaTables(postgres)
			if err != nil {
				logger.Error("Error checking core schema after migrations", "error", err)
				return nil, err
			}
			if !hasCoreSchema {
				logger.Info("Core schema tables are missing after SQL migrations, bootstrapping registration schema")
				if err := db.BootstrapRegistrationSchema(postgres); err != nil {
					logger.Error("Error bootstrapping registration schema", "error", err)
					return nil, err
				}
			}
		} else {
			logger.Info("No SQL migrations found, using GORM bootstrap for core schema")
		}
	}
	if conf.AppEnv == "DEV" || !hasSQLMigrations {
		if err := db.BootstrapRegistrationSchema(postgres); err != nil {
			logger.Error("Error bootstrapping registration schema", "error", err)
			return nil, err
		}
	}
	s3_storege, err := storage.NewSelectelS3(conf, logger)
	if err != nil {
		logger.Error("Error connect S3: %s", err)
	}
	logger.Info("S3 initialized")
	validator := validator.NewValidator(conf)

	popularEventsService, err := event.NewPopularEventsService(
		logger,
		postgres,
		redis,
		conf,
	)
	if err != nil {
		logger.Error("Error initializing popular events service", "error", err)
		return nil, err
	}
	logger.Info("Popular events service initialized")

	if err := popularEventsService.Start(); err != nil {
		logger.Error("Error starting popular events cron", "error", err)
		return nil, err
	}
	logger.Info("Popular events cron started")

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
	if errs := db.Seeder(postgres); len(errs) > 0 {
		for _, seedErr := range errs {
			logger.Error("Seeder failed", "error", seedErr)
		}
		return nil, errors.Join(errs...)
	}

	s := &Server{
		engine:               r,
		logger:               logger,
		postgres:             postgres,
		S3:                   s3_storege,
		mongo:                mongo,
		sessionStore:         sessionStore,
		cfg:                  *conf,
		validators:           validator,
		popularEventsService: popularEventsService,
	}
	s.initRouters()
	logger.Info("Server init")
	return s, nil
}

func (s *Server) Run(addr string) error {
	s.logger.Info("Server running on " + addr)
	return s.engine.Run(addr)
}

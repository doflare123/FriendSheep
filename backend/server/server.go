package server

import (
	storage "friendship/S3"
	"friendship/config"
	"friendship/db"
	"friendship/logger"
	"friendship/models"
	"friendship/models/events"
	"friendship/models/groups"
	statsusers "friendship/models/stats_users"
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
	if conf.AppEnv == "DEV" {
		gin.SetMode(gin.DebugMode)
		if err := db.AutoMigDB(postgres, &events.Event{}, &events.AgeLimit{}, &events.EventLocation{}, &events.Status{}, &events.EventsUser{}, &events.Genre{}, &events.EventGenre{}, &events.EventGenre{},
			&statsusers.Genre{}, statsusers.PopSessionType{}, statsusers.SettingTile{}, &models.User{}, models.StatsProcessedEvent{},
			&groups.Group{}, &groups.GroupContact{}, &groups.GroupGroupCategory{}, &models.Category{}, &groups.GroupUsers{}, &groups.GroupJoinRequest{}, &groups.GroupJoinInvite{}, &groups.GroupBlacklist{}, &groups.GroupActionLog{},
		); err != nil {
			logger.Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := db.MigrationDB(postgres, logger); err != nil {
			logger.Error("Error with migrations: %s", err)
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
	db.Seeder(postgres)

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

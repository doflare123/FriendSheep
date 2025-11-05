package server

import (
	"friendship/config"
	"friendship/db"
	"friendship/logger"
	"friendship/models"
	"friendship/repository"
	"log"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine   *gin.Engine
	logger   logger.Logger
	postgres repository.PostgresRepository
	mongo    repository.MongoRepository
	redis    repository.RedisRepository
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

	s := &Server{
		engine:   r,
		logger:   logger,
		postgres: postgres,
		mongo:    mongo,
		redis:    redis,
	}
	logger.Info("Server init")
	return s, nil
}

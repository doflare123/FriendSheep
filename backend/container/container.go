package container

import (
	"friendship/config"
	"friendship/logger"
	"friendship/repository"
)

type Container interface {
	GetConfig() *config.Config
	GetLogger() logger.Logger
	GetPostgres() repository.PostgresRepository
}

type container struct {
	config  *config.Config
	postres repository.PostgresRepository
	mongo   repository.MongoRepository
	redis   repository.RedisRepository
	logger  logger.Logger
}

func NewContainer(postres repository.PostgresRepository, mongo repository.MongoRepository, redis repository.RedisRepository, logger logger.Logger, conf config.Config) (Container, error) {
	return &container{
		config:  &conf,
		postres: postres,
		mongo:   mongo,
		redis:   redis,
		logger:  logger,
	}, nil
}

func (c *container) GetConfig() *config.Config {
	return c.config
}

func (c *container) GetLogger() logger.Logger {
	return c.logger
}

func (c *container) GetPostgres() repository.PostgresRepository {
	return c.postres
}

func (c *container) GetMongo() repository.MongoRepository {
	return c.mongo
}

func (c *container) GetRedis() repository.RedisRepository {
	return c.redis
}

package server

import (
	"errors"
	"fmt"
	storage "friendship/S3"
	"friendship/config"
	"friendship/db"
	friendshipdocs "friendship/docs"
	"friendship/logger"
	"friendship/repository"
	event "friendship/services/events"
	session "friendship/sessions"
	"friendship/validator"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

func validateNonDevStartupWithSQLMigrations(startupMigrationsEnabled bool, hasCoreSchema bool, hasPendingJoinRequestIndex bool) error {
	if startupMigrationsEnabled {
		if !hasCoreSchema {
			return errors.New("core schema is incomplete after SQL migrations; aborting startup (non-DEV fail-fast)")
		}
		return nil
	}

	if !hasCoreSchema {
		return errors.New("startup SQL migrations are disabled and core schema is incomplete; set ENABLE_STARTUP_SQL_MIGRATIONS=true for planned rollout")
	}
	if !hasPendingJoinRequestIndex {
		return errors.New("startup SQL migrations are disabled and required migration contract is missing (idx_group_join_request_pending_unique); set ENABLE_STARTUP_SQL_MIGRATIONS=true and apply SQL migrations")
	}

	return nil
}

func validateNonDevStartupMigrationAssets(hasSQLMigrations bool) error {
	if hasSQLMigrations {
		return nil
	}

	return errors.New("startup aborted: SQL migration assets not found in non-DEV environment; GORM bootstrap fallback is disabled")
}

func discoverStartupMigrationAssets(appEnv string, discover func() (bool, error)) (bool, error) {
	if appEnv == "DEV" {
		return false, nil
	}
	return discover()
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
	hasSQLMigrations, err := discoverStartupMigrationAssets(conf.AppEnv, db.HasMigrationSource)
	if err != nil {
		logger.Error("Error checking migration source", "error", err)
		return nil, err
	}
	if conf.AppEnv == "DEV" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := validateNonDevStartupMigrationAssets(hasSQLMigrations); err != nil {
			return nil, err
		}

		if conf.EnableStartupSQLMigrations {
			if err := db.MigrationDB(postgres, logger); err != nil {
				logger.Error("Error with migrations: %s", err)
				return nil, err
			}
		}

		hasCoreSchema, err := db.HasCoreSchemaTables(postgres)
		if err != nil {
			logger.Error("Error checking core schema after startup migration contract", "error", err)
			return nil, err
		}
		hasPendingJoinRequestIndex, err := db.HasPendingJoinRequestUniqueIndex(postgres)
		if err != nil {
			logger.Error("Error checking pending join request unique index contract", "error", err)
			return nil, err
		}
		if err := validateNonDevStartupWithSQLMigrations(conf.EnableStartupSQLMigrations, hasCoreSchema, hasPendingJoinRequestIndex); err != nil {
			return nil, err
		}
		if !conf.EnableStartupSQLMigrations {
			logger.Info("Startup SQL migrations disabled; core schema is already present, continuing without MigrationDB")
		}
	}
	if conf.AppEnv == "DEV" {
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
	friendshipdocs.SwaggerInfo.Description = "Operational runbooks: /docs/runbooks"
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
	r.GET("/docs/runbooks", func(c *gin.Context) {
		migrationDir, err := db.ResolveMigrationDirectory()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if migrationDir == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<html><head><meta charset=\"utf-8\"><title>Runbooks</title></head><body><h1>Runbooks</h1><ul></ul><p><a href=\"/swagger/index.html\">Back to Swagger</a></p></body></html>"))
			return
		}

		entries, err := os.ReadDir(migrationDir)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read migration dir: %v", err)})
			return
		}

		var items []string
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.Contains(strings.ToLower(name), "runbook") && strings.HasSuffix(strings.ToLower(name), ".md") {
				items = append(items, name)
			}
		}

		var b strings.Builder
		b.WriteString("<html><head><meta charset=\"utf-8\"><title>Runbooks</title></head><body>")
		b.WriteString("<h1>Runbooks</h1><ul>")
		for _, item := range items {
			b.WriteString("<li><a href=\"/docs/runbooks/" + item + "\">" + item + "</a></li>")
		}
		b.WriteString("</ul><p><a href=\"/swagger/index.html\">Back to Swagger</a></p></body></html>")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(b.String()))
	})
	r.GET("/docs/runbooks/:name", func(c *gin.Context) {
		name := filepath.Base(c.Param("name"))
		if !strings.HasSuffix(strings.ToLower(name), ".md") || !strings.Contains(strings.ToLower(name), "runbook") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid runbook name"})
			return
		}
		migrationDir, err := db.ResolveMigrationDirectory()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if migrationDir == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "runbook not found"})
			return
		}
		fullPath := filepath.Join(migrationDir, name)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "runbook not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read runbook: %v", err)})
			return
		}
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", content)
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

package server

import (
	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services"
	events "friendship/services/events"
	group "friendship/services/groups"
	"friendship/services/register"
	"friendship/services/sub"
	"friendship/utils"
)

func (s *Server) initRouters() {
	// jwt
	jwtService := utils.NewJWTUtils(s.cfg.JWTSecretKey)
	jwtMiddleware := middlewares.NewAuthMiddleware(jwtService)

	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(s.postgres)

	//регистрация и авторизация
	authsrv := services.NewAuthService(s.logger, jwtService, s.postgres)
	authH := handlers.NewAuthHandler(authsrv)
	routes.RegisterAuthRoutes(s.engine, authH)
	regsrv := register.NewRegisterSrv(s.logger, s.sessionStore, s.postgres, s.cfg, jwtService)
	regH := handlers.NewRegisterHandler(regsrv)
	routes.RegisterRegRoutes(s.engine, regH)

	//саб функции
	imgsrv := sub.NewImgService(s.logger, s.S3, s.validators.Image)
	subH := handlers.NewSubHandler(imgsrv)
	routes.RegisterSubRoutes(s.engine, subH)

	//регистрация групп
	groupsrv := group.NewGroupService(s.logger, s.postgres)
	groupH := handlers.NewGroupHandler(groupsrv)
	routes.RegisterGroupsRoutes(s.engine, groupH, jwtMiddleware, groupRoleMiddleware)

	//регистрация событий
	eventsrv := events.NewEventsService(s.logger, s.postgres)
	eventsH := handlers.NewEventsHandler(eventsrv)
	routes.RegisterEventsRoutes(s.engine, eventsH, jwtMiddleware, groupRoleMiddleware)
}

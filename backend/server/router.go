package server

import (
	"friendship/handlers"
	"friendship/routes"
	"friendship/services"
	"friendship/services/register"
	"friendship/services/sub"
)

func (s *Server) initRouters() {
	//регистрация и авторизация
	authsrv := services.NewAuthService(s.logger, s.cfg, s.postgres)
	authH := handlers.NewAuthHandler(authsrv)
	routes.RegisterAuthRoutes(s.engine, authH)
	regsrv := register.NewRegisterSrv(s.logger, s.sessionStore, s.postgres, s.cfg)
	regH := handlers.NewRegisterHandler(regsrv)
	routes.RegisterRegRoutes(s.engine, regH)
	//саб функции
	imgsrv := sub.NewImgService(s.logger, s.S3, s.validators.Image)
	subH := handlers.NewSubHandler(imgsrv)
	routes.RegisterSubRoutes(s.engine, subH)
}

package server

import (
	"friendship/handlers"
	"friendship/routes"
	"friendship/services"
	"friendship/services/register"
)

func (s *Server) initRouters() {
	//регистрация и авторизация
	authsrv := services.NewAuthService(s.logger, s.cfg, s.postgres)
	authH := handlers.NewAuthHandler(authsrv)
	routes.RegisterAuthRoutes(s.engine, authH)
	regsrv := register.NewRegisterSrv(s.logger, s.sessionStore, s.postgres, s.cfg)
	regH := handlers.NewRegisterHandler(regsrv)
	routes.RegisterRegRoutes(s.engine, regH)
}

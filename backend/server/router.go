package server

import (
	"friendship/handlers"
	"friendship/routes"
	"friendship/services"
)

func (s *Server) InitRouters() {
	//регистрация и авторизация
	authsrv := services.NewAuthService(s.logger, s.cfg, s.postgres)
	authH := handlers.NewAuthHandler(authsrv)
	routes.RegisterAuthRoutes(s.engine, authH)
}

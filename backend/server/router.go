package server

import (
	"friendship/handlers"
	"friendship/middlewares"
	"friendship/routes"
	"friendship/services"
	events "friendship/services/events"
	group "friendship/services/groups"
	"friendship/services/references"
	"friendship/services/register"
	"friendship/services/sub"
	"friendship/utils"
)

func (s *Server) initRouters() {
	verificationKeys := map[string]string{}
	if s.cfg.JWTPreviousKeyID != "" && s.cfg.JWTPreviousSecretKey != "" {
		verificationKeys[s.cfg.JWTPreviousKeyID] = s.cfg.JWTPreviousSecretKey
	}
	jwtService := utils.NewJWTUtilsWithConfig(s.cfg.JWTSecretKey, utils.JWTConfig{
		Issuer:           s.cfg.Auth.Issuer,
		Audience:         s.cfg.Auth.Audience,
		KeyID:            s.cfg.JWTKeyID,
		VerificationKeys: verificationKeys,
		AccessTTL:        s.cfg.Auth.AccessTokenTTL,
		ClockSkew:        s.cfg.Auth.ClockSkew,
	})
	authSessionStore, err := services.NewRedisAuthSessionStore(s.redis, services.AuthSessionStoreConfig{
		RefreshTTL: s.cfg.Auth.RefreshTokenTTL,
	})
	if err != nil {
		s.logger.Fatal("Failed to create auth session store", "error", err)
	}
	jwtMiddleware := middlewares.NewAuthMiddleware(jwtService, authSessionStore)
	jwtMiddleware.SetRateLimiter(s.rateLimiter)

	groupRoleMiddleware := middlewares.NewGroupRoleMiddleware(s.postgres)

	//регистрация и авторизация
	authsrv := services.NewAuthService(s.logger, jwtService, s.postgres, authSessionStore)
	authH := handlers.NewAuthHandler(authsrv)
	routes.RegisterAuthRoutes(s.engine, authH, jwtMiddleware)
	registrationStore := register.NewGORMRegistrationStore(s.postgres)
	regsrv, err := register.NewRegisterSrv(s.logger, s.sessionStore, registrationStore, &s.cfg, authsrv)
	if err != nil {
		s.logger.Fatal("Failed to create register service", "error", err)
	}
	regH := handlers.NewRegisterHandler(regsrv)
	routes.RegisterRegRoutes(s.engine, regH)

	//саб функции
	imgsrv := sub.NewImgService(s.logger, s.S3, s.validators.Image)
	subH := handlers.NewSubHandler(imgsrv)
	routes.RegisterSubRoutes(s.engine, subH, jwtMiddleware)

	//регистрация групп
	groupsrv := group.NewGroupService(s.logger, group.NewGORMGroupRepository(s.postgres))
	groupH := handlers.NewGroupHandler(groupsrv)
	routes.RegisterGroupsRoutes(s.engine, groupH, jwtMiddleware, groupRoleMiddleware)
	routes.RegisterUserGroupsRoutes(s.engine, groupH, jwtMiddleware)

	//регистрация событий
	eventUnitOfWork := events.NewGORMEventUnitOfWork(s.postgres)
	eventMembershipService := events.NewEventMembershipService(s.logger, eventUnitOfWork)
	eventCommandService := events.NewEventCommandService(s.logger, eventUnitOfWork)
	eventReadStore := events.NewGORMEventReadStore(s.postgres)
	eventReadService := events.NewEventReadService(s.logger, eventReadStore)
	eventAdminReader := events.NewGORMEventAdminReader(s.postgres)
	eventAdminService := events.NewEventAdminService(s.logger, eventAdminReader, eventUnitOfWork)
	eventLifecycleStore := events.NewGORMEventLifecycleStore(s.postgres)
	eventLifecycleService := events.NewEventLifecycleService(eventLifecycleStore, nil)
	eventLifecycleH := handlers.NewEventLifecycleHandler(eventLifecycleService)
	popularEventsH := handlers.NewPopularEventsHandler(s.popularEventsService)
	eventsH := handlers.NewEventsHandler(handlers.EventsHandlerDependencies{
		Membership: eventMembershipService,
		Commands:   eventCommandService,
		Reads:      eventReadService,
		Admin:      eventAdminService,
	})
	routes.RegisterEventsRoutes(s.engine, eventsH, popularEventsH, jwtMiddleware, groupRoleMiddleware)
	routes.RegisterInternalEventLifecycleRoutes(
		s.engine,
		eventLifecycleH,
		middlewares.NewInternalTokenMiddleware(s.cfg.NotifyServiceToken),
	)

	referenceStore := references.NewGORMReferenceStore(s.postgres)
	referenceService := references.NewReferenceService(referenceStore)
	referencesH := handlers.NewReferencesHandler(referenceService)
	routes.RegisterReferencesRoutes(s.engine, referencesH)
}

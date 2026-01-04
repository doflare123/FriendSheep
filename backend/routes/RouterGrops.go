package routes

import (
	"friendship/handlers"
	"friendship/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterGroupsRoutes(
	router *gin.Engine,
	groupHandler handlers.GroupHandler,
	authMiddleware *middlewares.AuthMiddleware,
	groupRoleMiddleware *middlewares.GroupRoleMiddleware,
) {
	// Публичные эндпоинты
	groupsPublic := router.Group("/api/v2/groups")
	groupsPublic.Use(authMiddleware.RequireAuth())
	{
		// Получение информации о группе
		groupsPublic.GET("/:groupId", groupHandler.GetGroupDetails)
		// Создание группы
		groupsPublic.POST("", groupHandler.CreateGroup)

		// Вступление в группу
		groupsPublic.POST("/:groupId/join", groupHandler.JoinGroup)

		// Выход из группы
		groupsPublic.POST("/:groupId/leave", groupHandler.LeaveGroup)

		// Приглашения - принятие и отклонение
		groupsPublic.POST("/invites/:inviteId/accept", groupHandler.AcceptJoinInvite)
		groupsPublic.POST("/invites/:inviteId/reject", groupHandler.RejectJoinInvite)
		groupsPublic.POST("/requests/:requestId/approve", groupHandler.ApproveJoinRequest)
		groupsPublic.POST("/requests/:requestId/reject", groupHandler.RejectJoinRequest)
	}

	// Эндпоинты для операторов и админов
	groupsOperator := router.Group("/api/v2/groups")
	groupsOperator.Use(authMiddleware.RequireAuth())
	groupsOperator.Use(groupRoleMiddleware.RequireOperatorOrAdmin())
	{
		// Обновление группы
		groupsOperator.PUT("", groupHandler.UpdateGroup)

		// Управление заявками
		groupsOperator.POST("/:groupId/requests/approve-all", groupHandler.ApproveAllJoinRequests)
		groupsOperator.POST("/:groupId/requests/reject-all", groupHandler.RejectAllJoinRequests)

		// Получение данных
		groupsOperator.GET("/:groupId/blacklist", groupHandler.GetGroupBlacklist)
		groupsOperator.GET("/:groupId/requests", groupHandler.GetJoinRequests)

		// Управление участниками
		groupsOperator.DELETE("/:groupId/members/:userId", groupHandler.DeleteUserFromGroup)
		groupsOperator.DELETE("/:groupId/blacklist/:userId", groupHandler.RemoveFromBlacklist)

		// Отправка приглашений
		groupsOperator.POST("/invites", groupHandler.CreateJoinInvite)
	}

	// Эндпоинты только для админов
	groupsAdmin := router.Group("/api/v2/groups")
	groupsAdmin.Use(authMiddleware.RequireAuth())
	groupsAdmin.Use(groupRoleMiddleware.RequireAdmin()) // Раскомментировать для проверки роли через middleware
	{
		// Удаление группы
		groupsAdmin.DELETE("/:groupId", groupHandler.DeleteGroup)

		// Управление правами
		groupsAdmin.POST("/permissions/add", groupHandler.AddPermissions)
		groupsAdmin.POST("/permissions/remove", groupHandler.RemovePermissions)

		// Просмотр истории действий
		groupsAdmin.GET("/:groupId/actions", groupHandler.WatchRecentActions)
	}
}

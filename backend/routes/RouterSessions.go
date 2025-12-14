package routes

// func RouterSessions(r *gin.Engine) {
// 	GroupSession := r.Group("api/sessions")
// 	{
// 		GroupSession.GET("/genres", middlewares.JWTAuthMiddleware(), handlers.GetGenres)
// 		GroupSession.POST("/createSession", middlewares.JWTAuthMiddleware(), handlers.CreateSession)
// 		GroupSession.POST("/join", middlewares.JWTAuthMiddleware(), handlers.JoinToSession)
// 		GroupSession.DELETE("/sessions/:id", middlewares.JWTAuthMiddleware(), handlers.DeleteSession)
// 		GroupSession.DELETE("/:sessionId/leave", middlewares.JWTAuthMiddleware(), handlers.LeaveSessionHandler)
// 	}
// 	GroupSessionAdmin := r.Group("api/admin/sessions")
// 	{
// 		GroupSessionAdmin.PATCH("/:sessionId", middlewares.JWTAuthMiddleware(), handlers.UpdateSessionHandler)
// 	}
// }

package handlers

// import (
// 	"friendship/models"
// 	"friendship/services"
// 	"net/http"
// 	"strconv"

// 	"github.com/gin-gonic/gin"
// )

// // GetNotify godoc
// // @Summary      Получить уведомления пользователя
// // @Description  Возвращает список непросмотренных уведомлений и ожидающих приглашений в группы для текущего пользователя.
// // @Tags         Users inf
// // @Security     BearerAuth
// // @Produce      json
// // @Success      200  {object}  services.GetNotifyResponse "Список уведомлений и приглашений"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// // @Router       /api/users/notify [get]
// func GetNotify(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}

// 	email, ok := emailValue.(string)
// 	if !ok {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email в контексте"})
// 		return
// 	}

// 	if email == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "email не может быть пустым"})
// 		return
// 	}

// 	res, err := services.GetNotify(email)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, res)
// }

// // GetNotify godoc
// // @Summary      Есть ли уведомления для пользователя или нет
// // @Description  Возвращает true - есть, нет - false
// // @Tags         Users inf
// // @Security     BearerAuth
// // @Produce      json
// // @Success      200  {object}  bool "Есть ли уведомления"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// // @Router       /api/users/notify/inf [get]
// func GetNotifyInf(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}

// 	email, ok := emailValue.(string)
// 	if !ok {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email в контексте"})
// 		return
// 	}

// 	if email == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "email не может быть пустым"})
// 		return
// 	}

// 	res, err := services.GetNotifyInf(email)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, res)
// }

// // MarkNotificationViewed godoc
// // @Summary      Отметить уведомление как просмотренное
// // @Description  Помечает указанное уведомление как просмотренное для текущего пользователя.
// // @Tags         Users inf
// // @Accept       json
// // @Produce      json
// // @Security     BearerAuth
// // @Param        notification body object{id=uint} true "ID уведомления для отметки"
// // @Success      200  {object}  map[string]string "Уведомление успешно отмечено как просмотренное"
// // @Failure      400  {object}  map[string]string "Некорректный запрос или уведомление не найдено"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Router       /api/users/notifications/viewed [post]
// func MarkNotificationViewed(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}

// 	email, ok := emailValue.(string)
// 	if !ok || email == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный email"})
// 		return
// 	}

// 	user := models.User{}

// 	var req struct {
// 		ID uint `json:"id"`
// 	}
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный запрос"})
// 		return
// 	}

// 	if err := services.MarkNotificationViewed(req.ID, user.ID); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "уведомление просмотрено"})
// }

// // ApproveInvite godoc
// // @Summary      Принять приглашение в группу
// // @Description  Принимает приглашение в группу по его ID. Пользователь добавляется в участники группы.
// // @Tags         Users inf
// // @Produce      json
// // @Security     BearerAuth
// // @Param        id   path      int  true  "ID приглашения"
// // @Success      200  {object}  map[string]string "Приглашение успешно принято"
// // @Failure      400  {object}  map[string]string "Некорректный ID или ошибка обработки приглашения (например, уже обработано или нет доступа)"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Router       /api/users/invites/{id}/approve [put]
// func ApproveInvite(c *gin.Context) {
// 	// Достаём email из контекста
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}

// 	email, ok := emailValue.(string)
// 	if !ok || email == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email"})
// 		return
// 	}

// 	inviteIDStr := c.Param("id")
// 	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
// 		return
// 	}

// 	// Вызываем сервис
// 	if err := services.ApproveInviteByID(email, uint(inviteID)); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "приглашение принято"})
// }

// // RejectInvite godoc
// // @Summary      Отклонить приглашение в группу
// // @Description  Отклоняет приглашение в группу по его ID.
// // @Tags         Users inf
// // @Produce      json
// // @Security     BearerAuth
// // @Param        id   path      int  true  "ID приглашения"
// // @Success      200  {object}  map[string]string "Приглашение успешно отклонено"
// // @Failure      400  {object}  map[string]string "Некорректный ID или ошибка обработки приглашения (например, уже обработано или нет доступа)"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Router       /api/users/invites/{id}/reject [put]
// func RejectInvite(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}

// 	email, ok := emailValue.(string)
// 	if !ok || email == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат email"})
// 		return
// 	}

// 	inviteIDStr := c.Param("id")
// 	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
// 		return
// 	}

// 	if err := services.RejectInviteByID(email, uint(inviteID)); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "приглашение отклонено"})
// }

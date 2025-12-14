package handlers

import (
	"errors"
	group "friendship/services/groups"
	"friendship/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupHandler interface {
	CreateGroup(c *gin.Context)
}

type groupHandler struct {
	srv group.GroupsService
}

func NewGroupHandler(srv group.GroupsService) GroupHandler {
	return &groupHandler{
		srv: srv,
	}
}

type CreateGroupInputDoc struct {
	Name             string `json:"name" example:"Моя группа"`
	Description      string `json:"description" example:"Полное описание"`
	SmallDescription string `json:"smallDescription" example:"Краткое описание"`
	Image            string `json:"image" example:"https://example.com/image.png"`
	IsPrivate        string `json:"isPrivate" example:"true"`
	City             string `json:"city,omitempty" example:"Москва"`
	Categories       []uint `json:"categories" example:"1,2,3"` // <- просто строка, swag так принимает
}

type JoinGroupInputDoc struct {
	GroupID uint `json:"groupId" binding:"required"`
}

type JoinGroupResponseDoc struct {
	Message string `json:"message" example:"Заявка на вступление отправлена"`
	Joined  bool   `json:"joined" example:"false"`
}

// CreateGroup godoc
// @Summary      Создание группы
// @Description  Создает новую группу. Контакты передаются строкой в формате "название:ссылка, название:ссылка"
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.CreateGroupInput true "Данные для создания группы"
// @Success      201 {object} dto.GroupDto "Группа успешно создана"
// @Failure      400 {object} map[string]interface{} "Некорректные данные"
// @Failure      401 {object} map[string]string "Не авторизован"
// @Failure      500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router       /api/v2/groups/create [post]
func (h *groupHandler) CreateGroup(c *gin.Context) {
	idValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Не найден email в контексте",
		})
		return
	}
	id := idValue.(uint)

	var input group.CreateGroupInput
	if err := c.ShouldBind(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if input.Image == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Изображение группы обязательно",
		})
		return
	}

	groupDto, err := h.srv.CreateGroup(id, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Пользователь не найден",
			})
		case errors.Is(err, group.ErrCategoriesNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Некоторые категории не найдены",
			})
		case errors.Is(err, group.ErrInvalidInput):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		case errors.Is(err, group.ErrGroupCreation):
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Ошибка при создании группы",
				"details": err.Error(),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, groupDto)
}

// // GetGroupInf godoc
// // @Summary Получить информацию о группе
// // @Description Получает информацию о группе по ID, включая список участников, категории, контакты и сессии. Для приватных групп требуется членство.
// // @Tags groups
// // @Security BearerAuth
// // @Param groupId path int true "ID группы"
// // @Produce json
// // @Success 200 {object} services.GroupInf "Информация о группе"
// // @Failure 400 {object} map[string]string "Некорректный ID группы"
// // @Failure 401 {object} map[string]string "Пользователь не авторизован"
// // @Failure 403 {object} map[string]string "Доступ к приватной группе запрещен"
// // @Failure 404 {object} map[string]string "Группа не найдена"
// // @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// // @Router /api/groups/{groupId} [get]
// func GetGroupInf(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}
// 	email := emailValue.(string)
// 	groupIDStr := c.Param("groupId")
// 	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)

// 	result, err := services.GetGroupInf(&groupID, &email)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }

// // @Security BearerAuth
// // joinToGroup godoc
// // @Summary Присоединение к группе
// // @Description Новый пользователь присоединяется к группе
// // @Tags groups
// // @Accept  json
// // @Produce  json
// // @Param input body JoinGroupInputDoc true "Данные для присоединения к группе"
// // @Success 200 {object} JoinGroupResponseDoc "Успешный ответ: вступил или заявка отправлена"
// // @Failure 400 {object} map[string]interface{}
// // @Failure 401 {object} map[string]string
// // @Failure 500 {object} map[string]string
// // @Router /api/groups/joinToGroup [post]
// func JoinGroup(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}
// 	email := emailValue.(string)

// 	var input services.JoinGroupInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный json", "details": err.Error()})
// 		return
// 	}
// 	res, err := services.JoinGroup(email, input)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": res.Message,
// 		"joined":  res.Joined,
// 	})
// }

// // DeleteGroups godoc
// // @Summary Удаление группы
// // @Description Удаляет группу. Только администратор группы может удалить её. Удаляются также участники и заявки на вступление.
// // @Tags groups_admin
// // @Security BearerAuth
// // @Param groupId path int true "ID группы"
// // @Produce json
// // @Success 200 {object} map[string]string "Группа успешно удалена"
// // @Failure 400 {object} map[string]string "Некорректный ID группы"
// // @Failure 403 {object} map[string]string "Нет прав или группа не найдена"
// // @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// // @Router /api/admin/groups/{groupId} [delete]
// func DeleteGroups(c *gin.Context) {
// 	groupIDParam := c.Param("groupId")
// 	groupID, err := strconv.ParseUint(groupIDParam, 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID группы"})
// 		return
// 	}

// 	if err := services.DeleteGroup(uint(groupID)); err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Вы удалили группу"})
// }

// // UpdateGroupHandler godoc
// // @Summary Обновить информацию о группе
// // @Description Позволяет администратору группы изменить её данные, включая контакты. Контакты передаются строкой в формате "название:ссылка, название:ссылка". Чтобы удалить все контакты, передайте пустую строку.
// // @Tags groups_admin
// // @Security BearerAuth
// // @Accept json
// // @Produce json
// // @Param groupId path int true "ID группы"
// // @Param input body services.GroupUpdateInput true "Новые данные группы. Для контактов используйте формат: 'vk:https://vk.com/mygroup, tg:https://t.me/mygroup'"
// // @Success 200 {object} map[string]string "Группа успешно обновлена"
// // @Failure 400 {object} map[string]string "Ошибка валидации или некорректный ID"
// // @Failure 403 {object} map[string]string "Нет прав на редактирование"
// // @Failure 404 {object} map[string]string "Группа не найдена"
// // @Failure 500 {object} map[string]string "Внутренняя ошибка"
// // @Router /api/admin/groups/{groupId} [patch]
// func UpdateGroupHandler(c *gin.Context) {
// 	groupIDStr := c.Param("groupId")
// 	groupID, err := strconv.ParseUint(groupIDStr, 10, 64)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID группы"})
// 		return
// 	}

// 	var input services.GroupUpdateInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "ошибка в теле запроса: " + err.Error()})
// 		return
// 	}

// 	if err := services.UpdateGroup(uint(groupID), input); err != nil {
// 		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"message": "Группа успешно обновлена"})
// }

// // SentJoinRequests godoc
// // @Summary      Отправить приглашение пользователю в группу
// // @Description  Администратор группы отправляет приглашение (заявку) указанному пользователю на вступление в группу.
// // @Tags         groups_admin
// // @Security     BearerAuth
// // @Produce      json
// // @Param        group_id query int true "ID группы, в которую отправляется приглашение" example(1)
// // @Param        user_id query int true "ID пользователя, которому отправляется приглашение" example(2)
// // @Success      200  {object}  services.JoinGroupResult "Заявка на вступление отправлена"
// // @Failure      400  {object}  map[string]string "Некорректные параметры запроса"
// // @Failure      401  {object}  map[string]string "Пользователь не авторизован"
// // @Failure      403  {object}  map[string]string "Нет прав администратора"
// // @Failure      404  {object}  map[string]string "Пользователь или группа не найдены"
// // @Failure      500  {object}  map[string]string "Внутренняя ошибка сервера"
// // @Router       /api/admin/groups/requestsForUser [post]
// func SentJoinRequests(c *gin.Context) {
// 	emailValue, exists := c.Get("email")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "не найден email в контексте"})
// 		return
// 	}
// 	email := emailValue.(string)
// 	var input services.SentJoinRequestsReq
// 	if err := c.ShouldBindQuery(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректные параметры запроса: " + err.Error()})
// 		return
// 	}
// 	res, err := services.SentJoinRequests(email, input)
// 	if err != nil {
// 		errStr := err.Error()
// 		if strings.Contains(errStr, "не администратор") {
// 			c.JSON(http.StatusForbidden, gin.H{"error": errStr})
// 		} else if strings.Contains(errStr, "не найден") {
// 			c.JSON(http.StatusNotFound, gin.H{"error": errStr})
// 		} else {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": errStr})
// 		}
// 		return
// 	}
// 	c.JSON(http.StatusOK, res)
// }

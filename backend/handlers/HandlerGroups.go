package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	group "friendship/services/groups"
	"friendship/utils"

	"github.com/gin-gonic/gin"
)

type GroupHandler interface {
	CreateGroup(c *gin.Context)
	UpdateGroup(c *gin.Context)
	DeleteGroup(c *gin.Context)
	JoinGroup(c *gin.Context)
	LeaveGroup(c *gin.Context)
	GetGroupDetails(c *gin.Context)
	GetManagedGroups(c *gin.Context)
	GetSubscribedGroups(c *gin.Context)
	SearchGroups(c *gin.Context)

	// Управление заявками
	ApproveAllJoinRequests(c *gin.Context)
	RejectAllJoinRequests(c *gin.Context)
	ApproveJoinRequest(c *gin.Context)
	RejectJoinRequest(c *gin.Context)
	GetJoinRequests(c *gin.Context)

	// Управление правами
	AddPermissions(c *gin.Context)
	RemovePermissions(c *gin.Context)
	GetGroupBlacklist(c *gin.Context)

	// Управление участниками
	DeleteUserFromGroup(c *gin.Context)
	RemoveFromBlacklist(c *gin.Context)

	// Приглашения
	CreateJoinInvite(c *gin.Context)
	AcceptJoinInvite(c *gin.Context)
	RejectJoinInvite(c *gin.Context)

	// История действий
	WatchRecentActions(c *gin.Context)
}

type groupHandler struct {
	srv group.GroupsService
}

func NewGroupHandler(srv group.GroupsService) GroupHandler {
	return &groupHandler{
		srv: srv,
	}
}

type GroupUpdateRequest struct {
	GroupID          uint    `json:"groupId" binding:"required"`
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	SmallDescription *string `json:"smallDescription"`
	Image            *string `json:"image"`
	IsPrivate        *bool   `json:"isPrivate"`
	City             *string `json:"city"`
	Categories       []*uint `json:"categories"`
	Contacts         *string `json:"contacts"`
}

type CreateGroupRequest struct {
	Name             string `json:"name" binding:"required,min=5,max=40" example:"Board Game Club"`
	Description      string `json:"description" binding:"required,min=5,max=300" example:"Group for board game fans"`
	SmallDescription string `json:"smallDescription" binding:"required,min=5,max=50" example:"Play together"`
	Image            string `json:"image" binding:"required,url" example:"https://cdn.example.com/images/board-games.jpg"`
	IsPrivate        *bool  `json:"isPrivate" binding:"required" example:"false"`
	City             string `json:"city,omitempty" example:"Moscow"`
	Categories       []uint `json:"categories" binding:"required,min=1" example:"1,3,5"`
	Contacts         string `json:"contacts,omitempty" example:"vk:https://vk.com/mygroup, tg:https://t.me/mygroup"`
}

func (r CreateGroupRequest) toServiceInput() group.CreateGroupInput {
	categories := make([]*uint, 0, len(r.Categories))
	for i := range r.Categories {
		categoryID := r.Categories[i]
		categories = append(categories, &categoryID)
	}

	return group.CreateGroupInput{
		Name:             r.Name,
		Description:      r.Description,
		SmallDescription: r.SmallDescription,
		Image:            r.Image,
		IsPrivate:        r.IsPrivate,
		City:             r.City,
		Categories:       categories,
		Contacts:         r.Contacts,
	}
}

// GetGroupDetails godoc
// @Summary      Получить информацию о группе
// @Description  Возвращает полную информацию о группе с участниками и активными событиями
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} dto.GroupFullDto "Информация о группе"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Группа приватная, доступ запрещен"
// @Failure      404 {object} dto.ErrorResponse "Группа не найдена"
// @Router       /api/v2/groups/{groupId} [get]
func (h *groupHandler) GetGroupDetails(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	groupDto, err := h.srv.GetGroupDetails(c.Request.Context(), userID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrInvalidInput):
			utils.BadRequest(c, "Некорректные данные")
		case errors.Is(err, group.ErrGroupNotFound):
			utils.NotFound(c, "Группа не найдена")
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Доступ к приватной группе запрещен")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, groupDto)
}

// GetManagedGroups godoc
// @Summary      Получить группы, где пользователь админ или модератор
// @Description  Возвращает группы текущего пользователя, разделённые на блоки администратора и модератора
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} dto.ManagedGroupsDto "Группы текущего пользователя по ролям"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/users/me/groups/managed [get]
func (h *groupHandler) GetManagedGroups(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		utils.Unauthorized(c, "Требуется авторизация")
		return
	}

	managedGroups, err := h.srv.GetManagedGroups(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrInvalidInput):
			utils.BadRequest(c, "Некорректные данные")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, managedGroups)
}

// GetSubscribedGroups godoc
// @Summary      Получить подписки пользователя на группы
// @Description  Возвращает страницу групп, где текущий пользователь состоит с ролью обычного участника. Группы, где пользователь администратор или модератор, не включаются.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Номер страницы" default(1) minimum(1)
// @Param        limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Success      200 {object} dto.SubscribedGroupsResponseDto "Подписки текущего пользователя на группы"
// @Failure      400 {object} dto.ErrorResponse "Некорректные параметры пагинации"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/users/me/groups/subscriptions [get]
func (h *groupHandler) GetSubscribedGroups(c *gin.Context) {
	userID := c.GetUint("userID")
	if userID == 0 {
		utils.Unauthorized(c, "Требуется авторизация")
		return
	}

	page, err := parseGroupSubscriptionsPaginationValue(
		c.Query("page"),
		group.DefaultSubscribedGroupsPage,
		0,
	)
	if err != nil {
		utils.BadRequest(c, "Некорректный параметр page")
		return
	}

	limit, err := parseGroupSubscriptionsPaginationValue(
		c.Query("limit"),
		group.DefaultSubscribedGroupsLimit,
		group.MaxSubscribedGroupsLimit,
	)
	if err != nil {
		utils.BadRequest(c, "Некорректный параметр limit")
		return
	}

	result, err := h.srv.GetSubscribedGroups(c.Request.Context(), userID, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrInvalidInput):
			utils.BadRequest(c, "Некорректные параметры пагинации")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseGroupSubscriptionsPaginationValue(raw string, defaultValue int, maxValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || maxValue > 0 && value > maxValue {
		return 0, group.ErrInvalidInput
	}

	return value, nil
}

// SearchGroups godoc
// @Summary      Поиск групп
// @Description  Публичный поиск групп с фильтрацией, сортировкой и пагинацией. Авторизация не требуется. При валидной авторизации isSubscribed отражает членство текущего пользователя, для анонимного запроса он равен false. Приватные группы возвращаются только как безопасные карточки без состава участников, контактов и административных данных.
// @Tags         groups
// @Produce      json
// @Param        q query string false "Поиск по названию и описанию группы"
// @Param        categoryIds query []int false "ID категорий (OR), повтором параметра или CSV"
// @Param        isPrivate query bool false "Фильтр приватности"
// @Param        city query string false "Поиск по городу"
// @Param        sortBy query string false "Поле сортировки" Enums(createdAt,memberCount,name) default(createdAt)
// @Param        sortOrder query string false "Направление сортировки" Enums(asc,desc) default(desc)
// @Param        page query int false "Номер страницы" default(1) minimum(1) maximum(10000)
// @Param        limit query int false "Размер страницы" default(20) minimum(1) maximum(100)
// @Success      200 {object} dto.GroupSearchResponseDto "Страница групп"
// @Failure      400 {object} dto.ErrorResponse "Некорректные параметры поиска"
// @Failure      401 {object} dto.ErrorResponse "Передан невалидный или отозванный токен"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Failure      503 {object} dto.ErrorResponse "Сервис авторизации временно недоступен"
// @Router       /api/v2/groups/search [get]
func (h *groupHandler) SearchGroups(c *gin.Context) {
	userID := c.GetUint("userID")

	input, err := bindGroupSearchQuery(c)
	if err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	result, err := h.srv.SearchGroups(c.Request.Context(), userID, input)
	if err != nil {
		if errors.Is(err, group.ErrInvalidGroupSearchInput) {
			utils.BadRequest(c, err.Error())
		} else {
			utils.InternalError(c, "Не удалось выполнить поиск групп")
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

func bindGroupSearchQuery(c *gin.Context) (group.GroupSearchInput, error) {
	input := group.GroupSearchInput{
		Query:     strings.TrimSpace(c.Query("q")),
		City:      strings.TrimSpace(c.Query("city")),
		SortBy:    strings.TrimSpace(c.Query("sortBy")),
		SortOrder: strings.ToLower(strings.TrimSpace(c.Query("sortOrder"))),
	}

	var err error
	input.CategoryIDs, err = parseGroupSearchCategoryIDs(c)
	if err != nil {
		return input, err
	}
	input.IsPrivate, err = parseGroupSearchOptionalBool(c, "isPrivate")
	if err != nil {
		return input, err
	}
	input.Page, err = parseGroupSearchOptionalInt(c.Query("page"), group.DefaultGroupSearchPage, "page")
	if err != nil {
		return input, err
	}
	input.Limit, err = parseGroupSearchOptionalInt(c.Query("limit"), group.DefaultGroupSearchLimit, "limit")
	if err != nil {
		return input, err
	}
	if input.Page < 1 || input.Page > group.MaxGroupSearchPage || input.Limit < 1 || input.Limit > group.MaxGroupSearchLimit {
		return input, group.ErrInvalidGroupSearchInput
	}
	if input.SortBy != "" && input.SortBy != group.GroupSearchSortCreatedAt && input.SortBy != group.GroupSearchSortMemberCount && input.SortBy != group.GroupSearchSortName {
		return input, group.ErrInvalidGroupSearchInput
	}
	if input.SortOrder != "" && input.SortOrder != group.GroupSearchOrderAscending && input.SortOrder != group.GroupSearchOrderDescending {
		return input, group.ErrInvalidGroupSearchInput
	}

	return input, nil
}

func parseGroupSearchCategoryIDs(c *gin.Context) ([]uint, error) {
	values := make([]string, 0)
	for _, raw := range c.QueryArray("categoryIds") {
		if raw == "" {
			continue
		}
		for _, part := range strings.Split(raw, ",") {
			value := strings.TrimSpace(part)
			if value == "" {
				return nil, group.ErrInvalidGroupSearchInput
			}
			values = append(values, value)
		}
	}

	result := make([]uint, 0, len(values))
	seen := make(map[uint]struct{}, len(values))
	for _, value := range values {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil || parsed == 0 {
			return nil, group.ErrInvalidGroupSearchInput
		}
		id := uint(parsed)
		if _, exists := seen[id]; exists {
			return nil, group.ErrInvalidGroupSearchInput
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	if len(result) > 100 {
		return nil, group.ErrInvalidGroupSearchInput
	}
	return result, nil
}

func parseGroupSearchOptionalBool(c *gin.Context, name string) (*bool, error) {
	raw, exists := c.Request.URL.Query()[name]
	if !exists {
		return nil, nil
	}
	if len(raw) != 1 || strings.TrimSpace(raw[0]) == "" {
		return nil, group.ErrInvalidGroupSearchInput
	}
	var value bool
	switch strings.TrimSpace(raw[0]) {
	case "true":
		value = true
	case "false":
		value = false
	default:
		return nil, group.ErrInvalidGroupSearchInput
	}
	return &value, nil
}

func parseGroupSearchOptionalInt(raw string, defaultValue int, name string) (int, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%w: %s", group.ErrInvalidGroupSearchInput, name)
	}
	return value, nil
}

// CreateGroup godoc
// @Summary      Создание группы
// @Description  Создает новую группу. Контакты передаются строкой в формате "название:ссылка, название:ссылка"
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateGroupRequest true "Данные для создания группы"
// @Success      201 {object} dto.GroupFullDto "Группа успешно создана"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/groups [post]
func (h *groupHandler) CreateGroup(c *gin.Context) {
	idValue, exists := c.Get("userID")
	if !exists {
		utils.Unauthorized(c, "Не найден userID в контексте")
		return
	}
	id := idValue.(uint)

	var request CreateGroupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ValidationError(c, err)
		return
	}

	if request.Image == "" {
		utils.BadRequest(c, "Изображение группы обязательно")
		return
	}

	groupDto, err := h.srv.CreateGroup(c.Request.Context(), id, request.toServiceInput())
	if err != nil {
		switch {
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrCategoriesNotFound):
			utils.BadRequest(c, "Некоторые категории не найдены")
		case errors.Is(err, group.ErrInvalidInput):
			utils.BadRequest(c, err.Error())
		case errors.Is(err, group.ErrGroupCreation):
			utils.InternalError(c, "Ошибка при создании группы", utils.WithDetails(err.Error()))
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, groupDto)
}

// UpdateGroup godoc
// @Summary      Обновление группы
// @Description  Обновляет информацию о группе. Все поля, кроме идентификатора группы, опциональны
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body GroupUpdateRequest true "Данные для обновления группы"
// @Success      200 {object} dto.GroupFullDto "Группа успешно обновлена"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Группа не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /api/v2/groups [put]
func (h *groupHandler) UpdateGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	var request GroupUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ValidationError(c, err)
		return
	}

	// Преобразуем request в input для сервиса
	input := group.GroupUpdateInput{
		GroupID:          request.GroupID,
		Name:             request.Name,
		Description:      request.Description,
		SmallDescription: request.SmallDescription,
		Image:            request.Image,
		IsPrivate:        request.IsPrivate,
		City:             request.City,
		Categories:       request.Categories,
		Contacts:         request.Contacts,
	}

	groupDto, err := h.srv.UpdateGroup(c.Request.Context(), actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			utils.NotFound(c, "Группа не найдена")
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, groupDto)
}

// DeleteGroup godoc
// @Summary      Удаление группы
// @Description  Удаляет группу. Доступно только для админа группы
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Группа успешно удалена"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Группа не найдена"
// @Router       /api/v2/groups/{groupId} [delete]
func (h *groupHandler) DeleteGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	success, err := h.srv.DeleteGroup(c.Request.Context(), actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			utils.NotFound(c, "Группа не найдена")
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Только админ может удалить группу")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Группа успешно удалена",
	})
}

// JoinGroup godoc
// @Summary      Вступление в группу
// @Description  Для открытых групп - сразу добавляет пользователя. Для приватных - создает заявку на вступление
// @Tags         groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} group.GroupResult "Результат вступления в группу"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      404 {object} dto.ErrorResponse "Группа не найдена"
// @Failure      409 {object} dto.ErrorResponse "Уже в группе или заявка уже отправлена"
// @Router       /api/v2/groups/{groupId}/join [post]
func (h *groupHandler) JoinGroup(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	result, err := h.srv.JoinGroup(c.Request.Context(), userID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrGroupNotFound):
			utils.NotFound(c, "Группа не найдена")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrAlreadyInGroup):
			utils.JSONError(c, http.StatusConflict, "Вы уже состоите в этой группе")
		case errors.Is(err, group.ErrRequestAlreadyExists):
			utils.JSONError(c, http.StatusConflict, "Ваша заявка уже отправлена")
		case errors.Is(err, group.ErrUserInBlacklist):
			utils.Forbidden(c, "Вы в черном списке этой группы")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// LeaveGroup godoc
// @Summary      Выход из группы
// @Description  Пользователь покидает группу. Админ не может покинуть группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Вы покинули группу"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Админ не может покинуть группу"
// @Failure      404 {object} dto.ErrorResponse "Вы не состоите в группе"
// @Router       /api/v2/groups/{groupId}/leave [post]
func (h *groupHandler) LeaveGroup(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	success, err := h.srv.LeaveGroup(c.Request.Context(), userID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotInGroup):
			utils.NotFound(c, "Вы не состоите в этой группе")
		default:
			utils.Forbidden(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Вы успешно покинули группу",
	})
}

// ApproveAllJoinRequests godoc
// @Summary      Одобрить все заявки на вступление
// @Description  Одобряет все ожидающие заявки на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Количество одобренных заявок"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests/approve-all [post]
func (h *groupHandler) ApproveAllJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	count, err := h.srv.ApproveAllJoinRequests(c.Request.Context(), actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявки одобрены",
		"count":   count,
	})
}

// RejectAllJoinRequests godoc
// @Summary      Отклонить все заявки на вступление
// @Description  Отклоняет все ожидающие заявки на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Success      200 {object} map[string]interface{} "Количество отклоненных заявок"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests/reject-all [post]
func (h *groupHandler) RejectAllJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	count, err := h.srv.RejectAllJoinRequests(c.Request.Context(), actorID, uint(groupID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Заявки отклонены",
		"count":   count,
	})
}

// ApproveJoinRequest godoc
// @Summary      Одобрить заявку на вступление
// @Description  Одобряет конкретную заявку на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        requestId path int true "ID заявки"
// @Success      200 {object} map[string]interface{} "Заявка одобрена"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Заявка не найдена"
// @Router       /api/v2/groups/requests/{requestId}/approve [post]
func (h *groupHandler) ApproveJoinRequest(c *gin.Context) {
	actorID := c.GetUint("userID")

	requestIDStr := c.Param("requestId")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID заявки")
		return
	}

	success, err := h.srv.ApproveJoinRequest(c.Request.Context(), actorID, uint(requestID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrUserInBlacklist):
			utils.Forbidden(c, "Пользователь в черном списке")
		case errors.Is(err, group.ErrJoinRequestNotFound):
			utils.NotFound(c, "Заявка не найдена")
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Заявка одобрена",
	})
}

// RejectJoinRequest godoc
// @Summary      Отклонить заявку на вступление
// @Description  Отклоняет конкретную заявку на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        requestId path int true "ID заявки"
// @Success      200 {object} map[string]interface{} "Заявка отклонена"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Заявка не найдена"
// @Router       /api/v2/groups/requests/{requestId}/reject [post]
func (h *groupHandler) RejectJoinRequest(c *gin.Context) {
	actorID := c.GetUint("userID")

	requestIDStr := c.Param("requestId")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID заявки")
		return
	}

	success, err := h.srv.RejectJoinRequest(c.Request.Context(), actorID, uint(requestID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrJoinRequestNotFound):
			utils.NotFound(c, "Заявка не найдена")
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Заявка отклонена",
	})
}

// AddPermissions godoc
// @Summary      Назначить оператора
// @Description  Назначает пользователя оператором группы. Доступно только для админа
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.GroupUserInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Права успешно назначены"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Router       /api/v2/groups/permissions/add [post]
func (h *groupHandler) AddPermissions(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.GroupUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.AddPermissions(c.Request.Context(), actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Только админ может назначать операторов")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrNotInGroup):
			utils.BadRequest(c, "Пользователь не состоит в группе")
		case errors.Is(err, group.ErrCannotChangeOwnRole):
			utils.BadRequest(c, "Нельзя изменить собственную роль")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь назначен оператором группы",
	})
}

// RemovePermissions godoc
// @Summary      Снять права оператора
// @Description  Снимает с пользователя права оператора группы. Доступно только для админа
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.GroupUserInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Права успешно сняты"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Router       /api/v2/groups/permissions/remove [post]
func (h *groupHandler) RemovePermissions(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.GroupUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.RemovePermissions(c.Request.Context(), actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Только админ может снимать права оператора")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrNotInGroup):
			utils.BadRequest(c, "Пользователь не состоит в группе")
		case errors.Is(err, group.ErrCannotChangeOwnRole):
			utils.BadRequest(c, "Нельзя изменить собственную роль")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Права оператора сняты",
	})
}

// DeleteUserFromGroup godoc
// @Summary      Удалить пользователя из группы
// @Description  Удаляет пользователя из группы и добавляет в черный список. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        userId path int true "ID пользователя"
// @Success      200 {object} map[string]interface{} "Пользователь удален из группы"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Router       /api/v2/groups/{groupId}/members/{userId} [delete]
func (h *groupHandler) DeleteUserFromGroup(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID пользователя")
		return
	}

	success, err := h.srv.DeleteUserFromGroup(c.Request.Context(), actorID, uint(groupID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrNotInGroup):
			utils.NotFound(c, "Пользователь не состоит в группе")
		case errors.Is(err, group.ErrCannotRemoveSelf):
			utils.BadRequest(c, "Нельзя удалить самого себя")
		default:
			utils.BadRequest(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь удален из группы и добавлен в черный список",
	})
}

// RemoveFromBlacklist godoc
// @Summary      Убрать из черного списка
// @Description  Удаляет пользователя из черного списка группы. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        userId path int true "ID пользователя"
// @Success      200 {object} map[string]interface{} "Пользователь удален из черного списка"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден в черном списке"
// @Router       /api/v2/groups/{groupId}/blacklist/{userId} [delete]
func (h *groupHandler) RemoveFromBlacklist(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	userIDStr := c.Param("userId")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID пользователя")
		return
	}

	success, err := h.srv.RemoveFromBlacklist(c.Request.Context(), actorID, uint(groupID), uint(targetUserID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		default:
			utils.NotFound(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Пользователь удален из черного списка",
	})
}

// CreateJoinInvite godoc
// @Summary      Создать приглашение в группу
// @Description  Отправляет приглашение пользователю вступить в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body group.GroupUserInput true "ID группы и пользователя"
// @Success      200 {object} map[string]interface{} "Приглашение отправлено"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Пользователь не найден"
// @Failure      409 {object} dto.ErrorResponse "Приглашение уже существует или пользователь уже в группе"
// @Router       /api/v2/groups/invites [post]
func (h *groupHandler) CreateJoinInvite(c *gin.Context) {
	actorID := c.GetUint("userID")

	var input group.GroupUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ValidationError(c, err)
		return
	}

	success, err := h.srv.CreateJoinInvite(c.Request.Context(), actorID, input)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrUserNotFound):
			utils.NotFound(c, "Пользователь не найден")
		case errors.Is(err, group.ErrAlreadyInGroup):
			utils.JSONError(c, http.StatusConflict, "Пользователь уже в группе")
		case errors.Is(err, group.ErrInviteAlreadyExists):
			utils.JSONError(c, http.StatusConflict, "Приглашение уже отправлено")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Приглашение отправлено",
	})
}

// AcceptJoinInvite godoc
// @Summary      Принять приглашение в группу
// @Description  Принимает приглашение и добавляет пользователя в группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения"
// @Success      200 {object} group.GroupResult "Приглашение принято"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Приглашение не найдено"
// @Router       /api/v2/groups/invites/{inviteId}/accept [post]
func (h *groupHandler) AcceptJoinInvite(c *gin.Context) {
	userID := c.GetUint("userID")

	inviteIDStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID приглашения")
		return
	}

	result, err := h.srv.AcceptJoinInvite(c.Request.Context(), userID, uint(inviteID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrUserInBlacklist):
			utils.Forbidden(c, "Вы в черном списке этой группы")
		case errors.Is(err, group.ErrInviteNotFound):
			utils.NotFound(c, "Приглашение не найдено")
		case errors.Is(err, group.ErrInviteNotOwned):
			utils.Forbidden(c, "Это приглашение вам не принадлежит")
		case errors.Is(err, group.ErrInviteAlreadyHandled):
			utils.BadRequest(c, "Приглашение уже обработано")
		default:
			utils.InternalError(c, "Ошибка при принятии приглашения", utils.WithDetails(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// RejectJoinInvite godoc
// @Summary      Отклонить приглашение в группу
// @Description  Отклоняет приглашение вступить в группу
// @Tags         groups
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения"
// @Success      200 {object} map[string]interface{} "Приглашение отклонено"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Failure      404 {object} dto.ErrorResponse "Приглашение не найдено"
// @Router       /api/v2/groups/invites/{inviteId}/reject [post]
func (h *groupHandler) RejectJoinInvite(c *gin.Context) {
	userID := c.GetUint("userID")

	inviteIDStr := c.Param("inviteId")
	inviteID, err := strconv.ParseUint(inviteIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID приглашения")
		return
	}

	success, err := h.srv.RejectJoinInvite(c.Request.Context(), userID, uint(inviteID))
	if err != nil {
		switch {
		case errors.Is(err, group.ErrInviteNotFound):
			utils.NotFound(c, "Приглашение не найдено")
		case errors.Is(err, group.ErrInviteNotOwned):
			utils.Forbidden(c, "Это приглашение вам не принадлежит")
		case errors.Is(err, group.ErrInviteAlreadyHandled):
			utils.BadRequest(c, "Приглашение уже обработано")
		default:
			utils.InternalError(c, "Ошибка при отклонении приглашения", utils.WithDetails(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Приглашение отклонено",
	})
}

// WatchRecentActions godoc
// @Summary      Получить историю действий в группе
// @Description  Возвращает список последних действий операторов и админов в группе (только админ)
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Param        action query string false "Код действия из справочника groupActionTypes"
// @Param        actionTypeId query int false "ID действия из справочника groupActionTypes"
// @Param        order query string false "Порядок createdAt: desc по умолчанию, asc для старых к новым"
// @Success      200 {array} group.GroupAction "История действий"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/actions [get]
func (h *groupHandler) WatchRecentActions(c *gin.Context) {
	userID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	var actionTypeID uint
	actionTypeIDStr := c.Query("actionTypeId")
	if actionTypeIDStr != "" {
		parsedActionTypeID, err := strconv.ParseUint(actionTypeIDStr, 10, 32)
		if err != nil {
			utils.BadRequest(c, "Некорректный ID типа действия")
			return
		}
		actionTypeID = uint(parsedActionTypeID)
	}

	actions, err := h.srv.WatchRecentActions(c.Request.Context(), userID, uint(groupID), group.GroupActionFilter{
		Limit:        limit,
		Action:       strings.TrimSpace(c.Query("action")),
		ActionTypeID: actionTypeID,
		Order:        strings.ToLower(strings.TrimSpace(c.DefaultQuery("order", "desc"))),
	})
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, actions)
}

// GetGroupBlacklist godoc
// @Summary      Получить черный список группы
// @Description  Возвращает список забаненных пользователей. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Success      200 {array} group.BlacklistUser "Черный список"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/blacklist [get]
func (h *groupHandler) GetGroupBlacklist(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	blacklist, err := h.srv.GetGroupBlacklist(c.Request.Context(), actorID, uint(groupID), limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, blacklist)
}

// GetJoinRequests godoc
// @Summary      Получить заявки на вступление
// @Description  Возвращает список заявок на вступление в группу. Доступно для админов и операторов
// @Tags         groups_admin
// @Produce      json
// @Security     BearerAuth
// @Param        groupId path int true "ID группы"
// @Param        status query string false "Статус заявки (pending, approved, rejected)"
// @Param        limit query int false "Количество записей (по умолчанию 50, максимум 100)"
// @Success      200 {array} group.JoinRequestInfo "Список заявок"
// @Failure      400 {object} dto.ErrorResponse "Некорректные данные"
// @Failure      401 {object} dto.ErrorResponse "Не авторизован"
// @Failure      403 {object} dto.ErrorResponse "Недостаточно прав"
// @Router       /api/v2/groups/{groupId}/requests [get]
func (h *groupHandler) GetJoinRequests(c *gin.Context) {
	actorID := c.GetUint("userID")

	groupIDStr := c.Param("groupId")
	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		utils.BadRequest(c, "Некорректный ID группы")
		return
	}

	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	requests, err := h.srv.GetJoinRequests(c.Request.Context(), actorID, uint(groupID), status, limit)
	if err != nil {
		switch {
		case errors.Is(err, group.ErrPermissionDenied):
			utils.Forbidden(c, "Недостаточно прав")
		case errors.Is(err, group.ErrNotInGroup):
			utils.Forbidden(c, "Вы не состоите в этой группе")
		default:
			utils.InternalError(c, err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, requests)
}

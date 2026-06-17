package groups

import "gorm.io/gorm"

type GroupActionType struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Code string `gorm:"uniqueIndex;not null" json:"code"`
	Name string `gorm:"not null" json:"name"`
}

const (
	ActionCreateGroup        = "create_group"
	ActionUpdateGroup        = "update_group"
	ActionJoinGroup          = "join_group"
	ActionCreateJoinRequest  = "create_join_request"
	ActionLeaveGroup         = "leave_group"
	ActionSendInvite         = "send_invite"
	ActionAcceptInvite       = "accept_invite"
	ActionRejectInvite       = "reject_invite"
	ActionApproveRequest     = "approve_request"
	ActionRejectRequest      = "reject_request"
	ActionApproveAllRequests = "approve_all_requests"
	ActionRejectAllRequests  = "reject_all_requests"
	ActionAddOperator        = "add_operator"
	ActionRemoveOperator     = "remove_operator"
	ActionChangeRole         = "change_role"
	ActionBanUser            = "ban_user"
	ActionUnbanUser          = "unban_user"
	ActionCreateEvent        = "create_event"
	ActionUpdateEvent        = "update_event"
	ActionDeleteEvent        = "delete_event"
	ActionJoinEvent          = "join_event"
	ActionLeaveEvent         = "leave_event"
	ActionKickFromEvent      = "kick_from_event"
)

type ActionTypeLookup interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
}

func DefaultGroupActionTypes() []GroupActionType {
	return []GroupActionType{
		{Code: ActionCreateGroup, Name: "Создание группы"},
		{Code: ActionUpdateGroup, Name: "Изменение группы"},
		{Code: ActionJoinGroup, Name: "Вступление в группу"},
		{Code: ActionCreateJoinRequest, Name: "Создание заявки на вступление"},
		{Code: ActionLeaveGroup, Name: "Выход из группы"},
		{Code: ActionSendInvite, Name: "Создание приглашения"},
		{Code: ActionAcceptInvite, Name: "Принятие приглашения"},
		{Code: ActionRejectInvite, Name: "Отклонение приглашения"},
		{Code: ActionApproveRequest, Name: "Одобрение заявки"},
		{Code: ActionRejectRequest, Name: "Отклонение заявки"},
		{Code: ActionApproveAllRequests, Name: "Одобрение всех заявок"},
		{Code: ActionRejectAllRequests, Name: "Отклонение всех заявок"},
		{Code: ActionAddOperator, Name: "Назначение оператора"},
		{Code: ActionRemoveOperator, Name: "Снятие оператора"},
		{Code: ActionChangeRole, Name: "Изменение роли"},
		{Code: ActionBanUser, Name: "Удаление участника"},
		{Code: ActionUnbanUser, Name: "Удаление из черного списка"},
		{Code: ActionCreateEvent, Name: "Создание события"},
		{Code: ActionUpdateEvent, Name: "Изменение события"},
		{Code: ActionDeleteEvent, Name: "Удаление события"},
		{Code: ActionJoinEvent, Name: "Вступление в событие"},
		{Code: ActionLeaveEvent, Name: "Выход из события"},
		{Code: ActionKickFromEvent, Name: "Удаление из события"},
	}
}

func FindGroupActionTypeID(store ActionTypeLookup, code string) (uint, error) {
	var actionType GroupActionType
	if err := store.Where("code = ?", code).First(&actionType).Error; err != nil {
		return 0, err
	}

	return actionType.ID, nil
}

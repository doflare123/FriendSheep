package dto

type SuccessResponse struct {
	Message string      `json:"message" example:"Операция выполнена успешно"`
	Data    interface{} `json:"data,omitempty"`
}

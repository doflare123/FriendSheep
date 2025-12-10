package dto

type ErrorResponse struct {
	Error   string `json:"error" example:"invalid_request"`
	Message string `json:"message" example:"Некорректный формат запроса"`
}

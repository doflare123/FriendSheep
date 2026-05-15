package dto

type ErrorResponse struct {
	Error        string            `json:"error" example:"invalid_request"`
	Message      string            `json:"message,omitempty" example:"Некорректный формат запроса"`
	Details      any               `json:"details,omitempty"`
	Fields       map[string]string `json:"fields,omitempty"`
	RequiredRole []string          `json:"required_role,omitempty"`
	YourRole     string            `json:"your_role,omitempty"`
}

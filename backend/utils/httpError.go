package utils

import (
	"friendship/models/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorOption func(*dto.ErrorResponse)

func BuildErrorResponse(errorText string, opts ...ErrorOption) dto.ErrorResponse {
	response := dto.ErrorResponse{
		Error:   errorText,
		Message: errorText,
	}

	for _, opt := range opts {
		opt(&response)
	}

	if response.Message == "" {
		response.Message = errorText
	}

	return response
}

func JSONError(c *gin.Context, status int, errorText string, opts ...ErrorOption) {
	c.JSON(status, BuildErrorResponse(errorText, opts...))
}

func AbortJSONError(c *gin.Context, status int, errorText string, opts ...ErrorOption) {
	JSONError(c, status, errorText, opts...)
	c.Abort()
}

func BadRequest(c *gin.Context, errorText string, opts ...ErrorOption) {
	JSONError(c, http.StatusBadRequest, errorText, opts...)
}

func Unauthorized(c *gin.Context, errorText string, opts ...ErrorOption) {
	JSONError(c, http.StatusUnauthorized, errorText, opts...)
}

func Forbidden(c *gin.Context, errorText string, opts ...ErrorOption) {
	JSONError(c, http.StatusForbidden, errorText, opts...)
}

func NotFound(c *gin.Context, errorText string, opts ...ErrorOption) {
	JSONError(c, http.StatusNotFound, errorText, opts...)
}

func InternalError(c *gin.Context, errorText string, opts ...ErrorOption) {
	JSONError(c, http.StatusInternalServerError, errorText, opts...)
}

func WithMessage(message string) ErrorOption {
	return func(response *dto.ErrorResponse) {
		response.Message = message
	}
}

func WithDetails(details any) ErrorOption {
	return func(response *dto.ErrorResponse) {
		response.Details = details
	}
}

func WithFields(fields map[string]string) ErrorOption {
	return func(response *dto.ErrorResponse) {
		response.Fields = fields
	}
}

func WithRequiredRoles(requiredRoles []string) ErrorOption {
	return func(response *dto.ErrorResponse) {
		if len(requiredRoles) == 0 {
			return
		}

		response.RequiredRole = append([]string(nil), requiredRoles...)
	}
}

func WithYourRole(roleName string) ErrorOption {
	return func(response *dto.ErrorResponse) {
		response.YourRole = roleName
	}
}

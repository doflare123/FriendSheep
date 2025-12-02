package routes

import (
	"friendship/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterSubRoutes(r *gin.Engine, subH handlers.SubHandler) {
	sub := r.Group("api/v2/sub")
	{
		sub.POST("UploadImg", subH.ChangePhoto)
	}
}

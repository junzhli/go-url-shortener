package sign

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthCheckHandler(context *gin.Context) {
	context.Status(http.StatusOK)
}

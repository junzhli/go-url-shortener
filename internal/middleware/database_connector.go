package middleware

import (
	"github.com/gin-gonic/gin"
	"url-shortener/internal/database"
)

func GetDatabaseConnector(service database.MySQLService) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("db", service)
		context.Next()
	}
}

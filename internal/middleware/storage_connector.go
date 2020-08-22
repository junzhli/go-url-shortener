package middleware

import (
	"github.com/gin-gonic/gin"
	"url-shortener/internal/cache"
	"url-shortener/internal/database"
)

func GetDatabaseConnector(service database.MySQLService) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("db", service)
		context.Next()
	}
}

func GetCacheConnector(service cache.Redis) gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("cache", service)
		context.Set("cache-service", cache.NewService(service))
		context.Next()
	}
}

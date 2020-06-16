package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"path"
	"time"
	"url-shortener/internal/database"
	"url-shortener/internal/middleware"
	"url-shortener/internal/route/shortener"
	userUrls "url-shortener/internal/route/user/shortener"
	"url-shortener/internal/route/user/sign"
)

// Start server, return error if failed to start.
func SetupServer(db database.MySQLService, jwtKey []byte, baseUrl string, domain string, htmlTemplate string) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob(path.Join(htmlTemplate, "*.tmpl"))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{baseUrl},
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.GetDatabaseConnector(db))

	apiRouter := r.Group("/api")
	{
		userRouter := apiRouter.Group("/user")
		{
			userRouter.GET("/authCheck", middleware.UserAuthenticated(jwtKey), sign.AuthCheckHandler)

			signRouter := userRouter.Group("/sign")
			{
				redirectUrl := fmt.Sprintf("%v/api/user/sign/google/callback", baseUrl)
				sign.VarConfig(redirectUrl, domain)

				googleOauth := signRouter.Group("/google")
				{
					googleOauth.GET("/", sign.GoogleSignHandler)
					googleOauth.GET("/callback", sign.GoogleSignCallbackHandler(jwtKey, baseUrl))
				}

				signRouter.POST("/", sign.UserSignInHandler(jwtKey))
			}

			userRouter.POST("/signup", sign.UserSignUpHandler)

			shortenerRouter := userRouter.Group("/url")
			{
				shortenerRouter.GET("/list", middleware.UserAuthenticated(jwtKey), userUrls.GetShortenUrlsHandler)
				shortenerRouter.DELETE("/r/:shorten_url", middleware.UserAuthenticated(jwtKey), userUrls.RemoveShortenUrlHandler)
			}
		}

		shortenerRouter := apiRouter.Group("/shortener")
		{
			shortenerRouter.POST("/", middleware.UserAuthenticated(jwtKey), shortener.CreateShortenUrlHandler)
			shortenerRouter.GET("/r/:shorten_url", shortener.GetShortenUrlHandler)
		}
	}

	return r
}

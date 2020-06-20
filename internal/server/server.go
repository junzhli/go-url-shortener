package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"path"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/database"
	"url-shortener/internal/middleware"
	"url-shortener/internal/route/shortener"
	userUrls "url-shortener/internal/route/user/shortener"
	"url-shortener/internal/route/user/sign"
	"url-shortener/internal/service/mail"
)

type ServerOptions struct {
	Database                 database.MySQLService
	Cache                    cache.Redis
	JwtKey                   []byte
	UseHttps                 bool
	BaseUrl                  string
	Domain                   string
	HtmlTemplate             string
	GoogleOauthConf          sign.GoogleOauthConfig
	EmailVerificationIgnored bool
	EmailRequest             chan<- mail.SendEmailOptions
}

// Start server, return error if failed to start.
func SetupServer(options ServerOptions) *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob(path.Join(options.HtmlTemplate, "*.tmpl"))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{options.BaseUrl},
		AllowMethods:     []string{"GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.GetDatabaseConnector(options.Database))
	r.Use(middleware.GetCacheConnector(options.Cache))

	apiRouter := r.Group("/api")
	{
		userRouter := apiRouter.Group("/user")
		{
			userRouter.GET("/authCheck", middleware.UserAuthenticated(options.JwtKey), sign.AuthCheckHandler)

			signRouter := userRouter.Group("/sign")
			{
				options.GoogleOauthConf.RedirectUrl = fmt.Sprintf("%v/api/user/sign/google/callback", options.BaseUrl)
				sign.VarConfig(options.Domain, options.GoogleOauthConf)

				googleOauth := signRouter.Group("/google")
				{
					googleOauth.GET("/", sign.GoogleSignHandler(options.UseHttps))
					googleOauth.GET("/callback", sign.GoogleSignCallbackHandler(options.JwtKey, options.BaseUrl))
				}

				signRouter.POST("/", sign.UserSignInHandler(options.JwtKey))
			}

			userRouter.POST("/signup", sign.UserSignUpHandler(options.EmailRequest, options.EmailVerificationIgnored))
			userRouter.POST("/signup/complete", sign.UserSignUpCompletionHandler(options.EmailVerificationIgnored))

			shortenerRouter := userRouter.Group("/url")
			{
				shortenerRouter.GET("/list", middleware.UserAuthenticated(options.JwtKey), userUrls.GetShortenUrlsHandler)
				shortenerRouter.DELETE("/r/:shorten_url", middleware.UserAuthenticated(options.JwtKey), userUrls.RemoveShortenUrlHandler)
			}
		}

		shortenerRouter := apiRouter.Group("/shortener")
		{
			shortenerRouter.POST("/", middleware.UserAuthenticated(options.JwtKey), shortener.CreateShortenUrlHandler(options.Domain))
			shortenerRouter.GET("/r/:shorten_url", shortener.GetShortenUrlHandler)
		}
	}

	return r
}

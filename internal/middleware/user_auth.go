package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
)

func UserAuthenticated(jwtKey []byte) gin.HandlerFunc {
	return func(context *gin.Context) {
		jwtToken, err := context.Cookie("accessToken")
		if err != nil {
			log.Println("No accessToken found on cookie header")
			context.AbortWithStatusJSON(http.StatusUnauthorized, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v\n", jwtToken)
			}

			return jwtKey, nil
		})
		if err != nil {
			log.Printf("token validation failed\n")
			context.AbortWithStatusJSON(http.StatusUnauthorized, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Printf("claims validation failed\n")
			context.AbortWithStatusJSON(http.StatusUnauthorized, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		elapsed := time.Since(time.Unix(int64((claims["issued"]).(float64)), 0)).Seconds()
		if elapsed > 86400*7 { // expire after 7 days
			log.Printf("access token expired\n")
			context.AbortWithStatusJSON(http.StatusUnauthorized, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		db := context.Value("db").(database.MySQLService)
		user, err := db.GetUserWithEmail(claims["email"].(string))
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); ok {
				log.Printf("given email not found in database\n")
				context.AbortWithStatusJSON(http.StatusUnauthorized, server.NewResponseErrorWithMessage(server.AuthenticationError))
				return
			}
			log.Printf("Unable to query for given email in database\n")
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		context.Set("claims", claims)
		context.Set("user", user)

		context.Next()
	}
}

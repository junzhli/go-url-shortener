package middleware

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"url-shortener/internal/database"
)

func UserAuthenticated(jwtKey []byte) gin.HandlerFunc {
	return func(context *gin.Context) {
		jwtToken, err := context.Cookie("accessToken")
		if err != nil {
			fmt.Println("No accessToken found on cookie header")
			context.String(http.StatusUnauthorized, "User not logged in")
			context.Abort()
			return
		}

		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", jwtToken)
			}

			return jwtKey, nil
		})

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			log.Printf("claims validation failed")
			context.String(http.StatusUnauthorized, "invalid accessToken")
			context.Abort()
			return
		}

		db := context.Value("db").(database.MySQLService)
		user, err := db.GetUserWithEmail(claims["email"].(string))
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); ok {
				log.Printf("given email not found in database")
				context.String(http.StatusUnauthorized, "invalid accessToken")
				context.Abort()
				return
			}
			log.Printf("Unable to query for given email in database")
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		context.Set("claims", claims)
		context.Set("user", user)

		context.Next()
	}
}

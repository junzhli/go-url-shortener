package sign

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
	"url-shortener/internal/util"
)

type Auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserSignInHandler(jwtKey []byte) gin.HandlerFunc {
	return func(context *gin.Context) {
		body := context.Request.Body
		r, err := ioutil.ReadAll(body)
		if err != nil {
			log.Printf("Unable to read body properly | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var auth Auth
		err = json.Unmarshal(r, &auth)
		if err != nil {
			log.Printf("Unexpected json string: %v | Reason: %v\n", string(r), err)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.InvalidJSONStringError))
			return
		}
		if !util.CheckEmailIfValid(auth.Email) {
			log.Printf("Email is not valid: %v\n", auth.Email)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.EmailValidationError))
			return
		}
		// TODO: password validation and sanitation

		db := context.Value("db").(database.MySQLService)
		userInfo, err := db.GetUserWithEmail(strings.ToLower(auth.Email))
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); ok {
				log.Printf("This user not found in database")
				context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.AuthenticationError))
				return
			}

			log.Printf("Unable to query for user info in database | Reason: %v", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if userInfo.Type != "local" {
			log.Printf("This user doesn't belong to this login type: %v", userInfo.Type)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		if !util.CheckPasswordHash(auth.Password, userInfo.Password) {
			log.Printf("Password hash mismatch")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.AuthenticationError))
			return
		}

		unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"type":   userInfo.Type,
			"email":  strings.ToLower(userInfo.Email),
			"issued": time.Now().Unix(),
		})
		issuedToken, err := unsignedToken.SignedString(jwtKey)

		context.JSON(http.StatusOK, gin.H{
			"issueToken": issuedToken,
		})
	}
}

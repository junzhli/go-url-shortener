package sign

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
	"url-shortener/internal/util"
)

func UserSignUpHandler(context *gin.Context) {
	body := context.Request.Body
	r, err := ioutil.ReadAll(body)
	if err != nil {
		log.Printf("Unable to read body properly | Reason: %v", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var auth Auth
	err = json.Unmarshal(r, &auth)
	if err != nil {
		log.Printf("Unexpected json string: %v | Reason: %v", string(r), err)
		context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.InvalidJSONStringError))
		return
	}
	if !util.CheckEmailIfValid(auth.Email) {
		log.Printf("Email is not valid\n")
		context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.EmailValidationError))
		return
	}
	// TODO: password validation and sanitation
	// TODO: email verifier e.g. confirmation email

	db := context.Value("db").(database.MySQLService)
	_, err = db.GetUserWithEmail(strings.ToLower(auth.Email))
	if err == nil {
		log.Printf("This user is registered in database\n")
		context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.AlreadyRegisteredError))
		return
	}

	if _, ok := err.(database.RecordNotFoundError); !ok {
		log.Printf("Unable to query for user info in database | Reason: %v\n", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	uuid, err := util.NewUUID()
	if err != nil {
		log.Printf("Error occurred when generating uuid | Reason: %v\n", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	hashedPassword, err := util.HashPassword(auth.Password)
	if err != nil {
		log.Printf("Error occurred when hashing password | Reason: %v\n", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = db.CreateUser(database.User{
		UserID:   uuid,
		Email:    strings.ToLower(auth.Email),
		Type:     database.UserTypeLocal,
		Password: hashedPassword,
	})
	if err != nil {
		log.Printf("Error occurred when creating user in database | Reason: %v\n", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	context.String(http.StatusOK, "Registered successfully")
}

package sign

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
	"url-shortener/internal/service/mail"
	"url-shortener/internal/util"
)

type signUpCompletion struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func UserSignUpCompletionHandler(emailVerificationIgnored bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		body := context.Request.Body
		r, err := ioutil.ReadAll(body)
		if err != nil {
			log.Printf("Unable to read body properly | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var verification signUpCompletion
		err = json.Unmarshal(r, &verification)
		if err != nil {
			log.Printf("Unexpected json string: %v | Reason: %v\n", string(r), err)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.InvalidJSONStringError))
			return
		}

		if !util.CheckEmailIfValid(verification.Email) {
			log.Printf("Email is not valid\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.EmailValidationError))
			return
		}

		fmt.Printf("code %v", verification.Code)
		if !util.IsOnlySixDigits(verification.Code) {
			log.Printf("Code is not valid\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.CodeValidationError))
			return
		}

		c := context.Value("cache").(cache.Redis)
		ck := cacheKey{Email: strings.ToLower(verification.Email)}
		code, err := c.Get(ck.CodeKey())
		if err == redis.Nil {
			log.Printf("No relevant registration info: code found in cache\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.CodeValidationError))
			return
		}
		if err != nil {
			log.Printf("Error occurred when getting code in cache | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !emailVerificationIgnored && verification.Code != code {
			log.Printf("Code mismatch\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.CodeValidationError))
			return
		}

		hashedPassword, err := c.Get(ck.PasswordKey())
		if err == redis.Nil {
			log.Printf("No relevant registration info: password found in cache\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.CodeValidationError))
			return
		}
		if err != nil {
			log.Printf("Error occurred when getting password in cache | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		db := context.Value("db").(database.MySQLService)
		uuid, err := util.NewUUID()
		if err != nil {
			log.Printf("Error occurred when generating uuid | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		err = db.CreateUser(database.User{
			UserID:   uuid,
			Email:    strings.ToLower(verification.Email),
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
}

func UserSignUpHandler(emailRequest chan<- mail.SendEmailOptions, emailVerificationIgnored bool) gin.HandlerFunc {
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
			log.Printf("Email is not valid\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.EmailValidationError))
			return
		}

		db := context.Value("db").(database.MySQLService)
		_, err = db.GetUserWithEmail(strings.ToLower(auth.Email))
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); !ok {
				log.Printf("Unable to query for user info in database | Reason: %v\n", err)
				context.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		} else {
			log.Printf("This user is registered in database\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.AlreadyRegisteredError))
			return
		}

		if len(auth.Password) > 20 || len(auth.Password) < 6 {
			log.Printf("Password is too weak or too long\n")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.PasswordValidationError))
			return
		}
		// TODO: sanitation
		hashedPassword, err := util.HashPassword(auth.Password)
		if err != nil {
			log.Printf("Error occurred when hashing password | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		expiration := time.Minute * 10
		ck := cacheKey{Email: strings.ToLower(auth.Email)}
		cache := context.Value("cache").(cache.Redis)
		tx := cache.NewTx()
		tx.Set(ck.PasswordKey(), hashedPassword, expiration)

		_code, err := rand.Int(rand.Reader, big.NewInt(999999))
		if err != nil {
			log.Printf(fmt.Sprintf("Failure on generating verification code | Reason: %v\n", err))
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		code := fmt.Sprintf("%06d", _code.Uint64())
		tx.Set(ck.CodeKey(), code, expiration)
		_, err = tx.Exec()
		if err != nil {
			log.Printf(fmt.Sprintf("Failure on generating verification code | Reason: %v\n", err))
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if emailVerificationIgnored {
			log.Printf("Warning: Registration request accepted without email verification\n")
			context.String(http.StatusOK, "Registration request accepted (without email verification)")
			return
		}

		go func() {
			emailRequest <- mail.SendEmailOptions{
				To:      auth.Email,
				Subject: "Registration confirmation",
				Message: fmt.Sprintf("Your verification code is %v", code),
			}
		}()

		context.String(http.StatusOK, "Registration request accepted")
		return
	}
}

type cacheKey struct {
	Email string
}

func (c cacheKey) PasswordKey() string {
	return fmt.Sprintf("%v:email", c.Email)
}

func (c cacheKey) CodeKey() string {
	return fmt.Sprintf("%v:code", c.Email)
}

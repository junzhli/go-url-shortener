package sign

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
	"url-shortener/internal/util"
)

func GoogleSignHandler(context *gin.Context) {
	state := generateStateOauthCookie(context)
	url := oauthConf.AuthCodeURL(state)
	context.Redirect(http.StatusFound, url)
}

func GoogleSignCallbackHandler(jwtKey []byte) gin.HandlerFunc {
	return func(context *gin.Context) {
		oauthState, err := context.Cookie("oauthstate")
		if err != nil {
			log.Printf("Failed to fetch oauthState: %v\n", err)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}
		if oauthState != context.Query("state") {
			log.Printf("oauthState verification failed: oauthState=%v queryStr=%v\n", oauthState, context.Query("state"))
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}

		userOauthInfo, err := extractUserInfoFromGoogleToken(context.Query("code"))
		if err != nil {
			log.Printf("Unable to extract user info with code: %v | Reason: %v\n", context.Query("code"), err)
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

		_, err = db.GetUserWithEmail(strings.ToLower(userOauthInfo.Email))
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); ok {
				log.Printf("User not registered\n")
				err = db.CreateGoogleUser(database.User{
					UserID: uuid,
					Email:  strings.ToLower(userOauthInfo.Email),
					Type:   database.UserTypeGoogle,
				}, database.GoogleUser{
					UserID:     uuid,
					GoogleUUID: userOauthInfo.Sub,
				})
				if err != nil {
					log.Printf("Error occurred when creating user in database | Reason: %v", err)
					context.AbortWithStatus(http.StatusInternalServerError)
					return
				}
			} else {
				fmt.Printf("Unable to check whether user is registered | Reason: %v\n", err)
				context.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}
		log.Printf("User has registered\n")

		unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"type":   database.UserTypeGoogle,
			"email":  strings.ToLower(userOauthInfo.Email),
			"issued": time.Now().Unix(),
		})
		issuedToken, err := unsignedToken.SignedString(jwtKey)

		context.HTML(http.StatusOK, "google_oauth_callback.tmpl", gin.H{
			"token": issuedToken,
		})
	}
}

type googleOauthUserInfo struct {
	/**
	{
	  "sub": "110442343466079368834",
	  "name": "Jeremy Lee",
	  "given_name": "Jeremy",
	  "family_name": "Lee",
	  "picture": "https://lh3.googleusercontent.com/a-/AOh14Giu0CCL6vIY2chGppGI6terV0NjImOHIGknP_9TrQ",
	  "email": "junzhli@gmail.com",
	  "email_verified": true,
	  "locale": "en"
	}
	*/
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

func extractUserInfoFromGoogleToken(code string) (*googleOauthUserInfo, error) {
	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Println("Unable to get token from authorization code")
		return nil, err
	}

	client := oauthConf.Client(context.Background(), token)

	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Println("Unable to fetch user info from google apis")
		return nil, err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Unable to read response body\n")
		return nil, err
	}

	var userInfo googleOauthUserInfo
	if err = json.Unmarshal(content, &userInfo); err != nil {
		log.Printf("Unable to parse response body as json obj\n")
		return nil, err
	}

	return &userInfo, nil
}

func generateStateOauthCookie(context *gin.Context) string {
	expiration := time.Now().Add(365 * 24 * time.Hour).Second()
	random := make([]byte, 16)
	rand.Read(random)
	state := base64.URLEncoding.EncodeToString(random)
	context.SetCookie("oauthstate", state, expiration, "/", domain, false, true)
	return state
}

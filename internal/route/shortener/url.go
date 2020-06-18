package shortener

import (
	"crypto/rand"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	url2 "net/url"
	"time"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
	"url-shortener/internal/util"
)

type ShortenReq struct {
	URL string `json:"url"`
}

func GetShortenUrlHandler(context *gin.Context) {
	shortenUrl := context.Param("shorten_url")

	db := context.Value("db").(database.MySQLService)

	url, err := db.GetURLWithShortenURL(shortenUrl)
	if err != nil {
		if _, ok := err.(database.RecordNotFoundError); ok {
			log.Printf("Given url %s not found in database", shortenUrl)
			context.Status(http.StatusNotFound)
			return
		}
		log.Printf("Error occurred when querying for url %s | Reason: %s", shortenUrl, err)
		context.Status(http.StatusInternalServerError)
		return
	}

	context.Redirect(http.StatusMovedPermanently, url.OriginURL)
}

func CreateShortenUrlHandler(domain string) gin.HandlerFunc {
	return func(context *gin.Context) {
		/**
		{
			"url": "<your-url>"
		}
		*/
		body := context.Request.Body
		r, err := ioutil.ReadAll(body)
		if err != nil {
			log.Printf("Unable to read body properly | Reason: %v\n", err)
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		var sReq ShortenReq
		err = json.Unmarshal(r, &sReq)
		if err != nil {
			log.Printf("Unexpected json string: %v | Reason: %v\n", string(r), err)
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}
		if len(sReq.URL) == 0 {
			log.Printf("Empty url")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}
		u, err := url2.Parse(sReq.URL)
		if err != nil {
			log.Printf("Invalid url to get shorthand")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			log.Printf("Invalid scheme to get shorthand")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}
		if u.Hostname() == domain {
			log.Printf("Recursive resolves is not allowed")
			context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
			return
		}

		db := context.Value("db").(database.MySQLService)
		user := context.Value("user").(*database.User)
		url, err := db.GetURLIfExistsWithUser(*user, u.String())
		if err != nil {
			if _, ok := err.(database.RecordNotFoundError); ok {
				shorten, err := getRandomUniqueStr(big.NewInt(999999999), time.Now())
				if err != nil {
					log.Printf("Unable to gen random number properly")
					context.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				err = db.CreateURL(u.String(), shorten, *user)
				if err != nil {
					log.Printf("Unable to create entity for given url | Reason: %s", err)
					context.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				context.JSON(http.StatusOK, gin.H{
					"url": shorten,
				})
				return
			}

			log.Printf("Error occurred when querying for given origin url if non-absent")
			context.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		context.JSON(http.StatusOK, gin.H{
			"url": url.ShortenURL,
		})
	}
}

// seed1 is from bigInteger in range of 0 to given value, seed2 is from time stamp in the form of seconds
func getRandomUniqueStr(seed1 *big.Int, seed2 time.Time) (string, error) {
	// TODO: better to generate unique id with single instance of offline unique id generator behind exposed entry-point
	randomNumber, err := rand.Int(rand.Reader, seed1)
	if err != nil {
		return "", err
	}
	return util.Base62FromBase10(randomNumber.Uint64()) + util.Base62FromBase10(uint64(seed2.Nanosecond())), nil
}

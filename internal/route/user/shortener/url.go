package shortener

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"url-shortener/internal/database"
	server "url-shortener/internal/route/error"
)

type URLsResponse struct {
	Total uint64        `json:"total"`
	URLs  []URLResponse `json:"urls"`
}

type URLResponse struct {
	OriginURL  string `json:"origin_url"`
	ShortenURL string `json:"shorten_url"`
}

func GetShortenUrlsHandler(context *gin.Context) {
	paramOffset := context.DefaultQuery("offset", "0")
	paramLimit := context.DefaultQuery("limit", "100")

	offset, err := strconv.ParseUint(paramOffset, 10, 64)
	if err != nil {
		log.Printf("Unable to decode query parameter offset: %v | Reason: %v\n", paramOffset, err)
		context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
		return
	}
	limit, err := strconv.ParseUint(paramLimit, 10, 64)
	if err != nil {
		log.Printf("Unable to decode query parameter limit: %v | Reason: %v\n", paramLimit, err)
		context.AbortWithStatusJSON(http.StatusBadRequest, server.NewResponseErrorWithMessage(server.RequestError))
	}
	if limit < 1 {
		limit = 100
	}

	db := context.Value("db").(database.MySQLService)
	user := context.Value("user").(*database.User)
	total, urls, err := db.GetURLsWithUser(*user, offset, limit)
	if err != nil {
		log.Printf("Unable to query for user's urls | Reason: %v\n", err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		log.Printf("No record found in database")
		context.AbortWithStatus(http.StatusNotFound)
		return
	}

	resUrls := make([]URLResponse, len(urls))
	for i, url := range urls {
		resUrls[i] = URLResponse{
			OriginURL:  url.OriginURL,
			ShortenURL: url.ShortenURL,
		}
	}

	context.JSON(http.StatusOK, URLsResponse{
		Total: total,
		URLs:  resUrls,
	})
}

func RemoveShortenUrlHandler(context *gin.Context) {
	url := context.Param("shorten_url")
	fmt.Printf("%v", url)

	db := context.Value("db").(database.MySQLService)
	err := db.DeleteURL(url)
	if err != nil {
		log.Printf("Unable to delete entity %v in database | Reason: %v\n", url, err)
		context.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	context.Status(http.StatusOK)
}

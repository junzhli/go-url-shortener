package server_test

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/server"
)

var _ = Describe("Server APIs", func() {
	var (
		db                     database.MySQLService
		router                 *gin.Engine
		user1                  database.User
		user2                  database.User
		user1AccessTokenHeader string
		user1Url               string
		user1ShortenUrl        string
		user1InvalidUrl        string
	)

	BeforeEach(func() {
		// TODO:
		user1 = database.User{
			Email:    "test5@test5.com",
			Password: "123456",
		}

		user2 = database.User{
			Email:    "test.com",
			Password: "123456",
		}

		user1Url = "https://www.google.com"
		user1InvalidUrl = "hxx://xxx...com"

		env := config.ReadEnv()

		/**
		Database configuration
		*/
		dbConfig := database.Config{
			Username: env.DBUser,
			Password: env.DBPass,
			Host:     env.DBHost,
			Port:     env.DBPort,
		}
		_db, err := database.NewMySQLDatabase(dbConfig)
		if err != nil {
			log.Fatalf("Unable to set up database | Reason: %v", err)
		}
		db = _db
		jwtKey := []byte(env.JwtKey)

		router = server.SetupServer(db, jwtKey, env.BaseUrl.String(), strings.Split(env.BaseUrl.Host, ":")[0])
	})

	Context("Sign up with local account", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user1.Email, user1.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusBadRequest)))
		})

		It("should reject the request due to email format problem", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user2.Email, user2.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject the request as email field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"email": "",
				"password": "%v"
			}
			`, user2.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject the request as password field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": ""
			}
			`, user2.Email)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/signup", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Sign in with local account", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"email": "%v",
				"password": "%v"
			}
			`, user1.Email, user1.Password)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/sign/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			user1AccessTokenHeader = recorder.Header().Get("Set-Cookie")
		})

		It("should reject due to field problem", func() {
			payload := fmt.Sprintf(`
			{
				
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/user/sign/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Generate a shorten url", func() {
		It("should perform successfully", func() {
			payload := fmt.Sprintf(`
			{
				"url": "%v"
			}
			`, user1Url)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
			bodyReader := recorder.Body
			body, err := ioutil.ReadAll(bodyReader)
			Expect(err).ShouldNot(HaveOccurred())

			var response map[string]interface{}
			err = json.Unmarshal(body, &response)
			Expect(err).ShouldNot(HaveOccurred())
			url, ok := response["url"]
			Expect(ok).To(Equal(true))
			urlStr, ok := url.(string)
			Expect(ok).To(Equal(true))
			user1ShortenUrl = urlStr
		})

		It("should reject as url field is empty", func() {
			payload := fmt.Sprintf(`
			{
				"url": ""
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject due to invalid url", func() {
			payload := fmt.Sprintf(`
			{
				"url": "%v"
			}
			`, user1InvalidUrl)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})

		It("should reject due to authorized problem", func() {
			payload := fmt.Sprintf(`
			{
				"url": "https://facebook.com"
			}
			`)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/shortener/", strings.NewReader(payload))
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("Resolve a shorten url", func() {
		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/shortener/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should reject due to invalid request", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/shortener/r/12345678", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Resolve a shorten url", func() {
		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/api/shortener/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should reject due to invalid request", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/shortener/r/12345678", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("Get user's urls", func() {
		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/url/list", nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should reject due to authorized problem", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/url/list", nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("Delete user's shorten url", func() {
		It("should reject due to authorized problem", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/user/url/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("should perform successfully", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/user/url/r/%v", user1ShortenUrl), nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusOK))

			recorder = httptest.NewRecorder()
			req = httptest.NewRequest("GET", fmt.Sprintf("/api/shortener/r/%v", user1ShortenUrl), nil)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})

		It("should reject as the deletion request of shorten url doesn't exist", func() {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("DELETE", "/api/shortener/r/12345678", nil)
			req.Header.Set("Cookie", user1AccessTokenHeader)
			router.ServeHTTP(recorder, req)
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
		})
	})

})

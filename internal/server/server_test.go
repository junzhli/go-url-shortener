package server_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
	"url-shortener/internal/server"
)

var _ = Describe("Server routing", func() {
	var (
		db     database.MySQLService
		router *gin.Engine
		user1  database.User
		user2  database.User
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
		})
	})

	AfterSuite(func() {
		err := db.DeleteUser(user1)
		Expect(err).ShouldNot(HaveOccurred())
	})
})

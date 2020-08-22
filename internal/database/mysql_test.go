package database_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"url-shortener/internal/config"
	"url-shortener/internal/database"
)

var _ = Describe("GormService (default impl of MySQLService)", func() {
	var (
		db          database.MySQLService
		user1       database.User
		user2       database.User
		googleUser2 database.GoogleUser
		user3Email  string
		url1        string
		url1S       string
		url2        string
		url2S       string
		url3        string
		url3S       string
		url4        string
	)

	BeforeEach(func() {
		user1 = database.User{
			UserID:   "test-user-1",
			Email:    "test1@test.com",
			Type:     "local",
			Password: "$2a$14$J9me9P5IdEsLE2BMU9AV1.qxBBx62y/8WK2NWEmawgfVNtX4c0hf.", // password: 1234
		}
		user2 = database.User{
			UserID: "test-user-2",
			Email:  "xxx@gmail.com",
			Type:   "google",
		}
		googleUser2 = database.GoogleUser{
			UserID:     "test-user-2",
			GoogleUUID: "12345678",
		}
		user3Email = "xyz@xyz.com"
		url1 = "https://google.com"
		url1S = "2sDftfgG"
		url2 = "https://google.com"
		url2S = "s2D0tf"
		url3 = "https://facebook.com"
		url3S = "s4rf"
		url4 = "https://twitter.com"

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
		Expect(err).NotTo(HaveOccurred())
		db = _db
	})

	Describe("Create user", func() {
		Context("From local account", func() {
			It("should create user successfully", func() {
				err := db.CreateUser(user1)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("From Google account", func() {
			It("should create user successfully", func() {
				err := db.CreateGoogleUser(user2, googleUser2)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Create shorten url", func() {
		Context("From local account", func() {
			It("should perform successfully", func() {
				err := db.CreateURL(url1, url1S, user1)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("From Google account", func() {
			It("should perform successfully", func() {
				err := db.CreateURL(url2, url2S, user2)
				Expect(err).NotTo(HaveOccurred())

				err = db.CreateURL(url3, url3S, user2)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Resolve shorten url", func() {
		It("should resolve successfully", func() {
			_url1, err := db.GetURLWithShortenURL(url1S)
			Expect(err).NotTo(HaveOccurred())
			Expect(_url1).NotTo(Equal(nil))
			Expect(_url1.ShortenURL).To(Equal(url1S))
			Expect(_url1.Owner).To(Equal(user1.UserID))
			Expect(_url1.OriginURL).To(Equal(url1))

			_url2, err := db.GetURLWithShortenURL(url2S)
			Expect(err).NotTo(HaveOccurred())
			Expect(_url2).NotTo(Equal(nil))
			Expect(_url2.ShortenURL).To(Equal(url2S))
			Expect(_url2.Owner).To(Equal(user2.UserID))
			Expect(_url2.OriginURL).To(Equal(url2))

			_url3, err := db.GetURLWithShortenURL(url3S)
			Expect(err).NotTo(HaveOccurred())
			Expect(_url3).NotTo(Equal(nil))
			Expect(_url3.ShortenURL).To(Equal(url3S))
			Expect(_url3.Owner).To(Equal(user2.UserID))
			Expect(_url3.OriginURL).To(Equal(url3))
		})
	})

	Describe("Update shorten url info (e.g. update count)", func() {
		It("should update successfully", func() {
			_url1, err := db.GetURLWithShortenURL(url1S)
			var updatedCount int64 = 1500
			_url1.Count = updatedCount
			err = db.UpdateURL(_url1)
			Expect(err).NotTo(HaveOccurred())
			_url1, err = db.GetURLWithShortenURL(url1S)
			Expect(_url1.Count).To(Equal(updatedCount))
		})
	})

	Describe("Get record if exists", func() {
		It("should not exist", func() {
			_, err := db.GetURLIfExistsWithUser(user1, url4)
			Expect(err).To(HaveOccurred())
			_, ok := err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))

			_, err = db.GetURLIfExistsWithUser(user2, url4)
			Expect(err).To(HaveOccurred())
			_, ok = err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))
		})
	})

	Describe("Get user info with email", func() {
		It("should perform successfully and get exact the same one", func() {
			_user1, err := db.GetUserWithEmail(user1.Email)
			Expect(err).NotTo(HaveOccurred())
			Expect(user1 == *_user1).To(Equal(true))

			_user2, err := db.GetUserWithEmail(user2.Email)
			Expect(err).NotTo(HaveOccurred())
			Expect(user2 == *_user2).To(Equal(true))
		})

		It("should get nothing", func() {
			_, err := db.GetUserWithEmail(user3Email)
			Expect(err).To(HaveOccurred())
			_, ok := err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))
		})
	})

	Describe("Get user info with id", func() {
		It("should perform successfully and get exact the same one", func() {
			_user1, err := db.GetUserWithID(user1.UserID)
			Expect(err).NotTo(HaveOccurred())
			Expect(user1 == *_user1).To(Equal(true))

			_user2, err := db.GetUserWithID(user2.UserID)
			Expect(err).NotTo(HaveOccurred())
			Expect(user2 == *_user2).To(Equal(true))
		})
	})

	Describe("Delete user's url in database", func() {
		It("should perform successfully", func() {
			err := db.DeleteURL(url1S)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.GetURLWithShortenURL(url1S)
			Expect(err).To(HaveOccurred())
			_, ok := err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))

			err = db.DeleteURL(url2S)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.GetURLWithShortenURL(url2S)
			Expect(err).To(HaveOccurred())
			_, ok = err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))

			err = db.DeleteURL(url3S)
			Expect(err).NotTo(HaveOccurred())
			_, err = db.GetURLWithShortenURL(url3S)
			Expect(err).To(HaveOccurred())
			_, ok = err.(database.RecordNotFoundError)
			Expect(ok).To(Equal(true))
		})
	})

	AfterSuite(func() {
		err := db.DeleteUser(user1)
		Expect(err).NotTo(HaveOccurred())
		err = db.DeleteUser(user2)
		Expect(err).NotTo(HaveOccurred())
	})
})

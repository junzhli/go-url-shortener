package cache_test

import (
	"fmt"
	rs "github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
	. "url-shortener/internal/cache"
	"url-shortener/internal/config"
)

var _ = Describe("Redis (Service)", func() {
	var (
		cache          Redis
		cacheService   Service
		testShortenURL string
		testOrigURL    string
		keyExpire      time.Duration
	)

	BeforeEach(func() {
		testShortenURL = "test1"
		testOrigURL = "https://google.com"
		keyExpire = 5 * time.Second

		env := config.ReadEnv()

		/**
		Cache configuration
		*/
		_cache := New(&rs.Options{
			Addr:         fmt.Sprintf("%v:%v", env.RedisHost, env.RedisPort),
			Password:     env.RedisPassword,
			DB:           0,
			ReadTimeout:  time.Minute,
			WriteTimeout: time.Minute,
		})
		cache = _cache
		cacheService = NewService(cache)
	})

	Describe("Put shorten url", func() {
		Context("Set key-pair with value represented by string", func() {
			It("should successfully", func() {
				err := cacheService.PutCachedURL(testShortenURL, testOrigURL, keyExpire)
				Expect(err).NotTo(HaveOccurred())
				_oriURL, err := cacheService.GetCachedURL(testShortenURL)
				Expect(err).NotTo(HaveOccurred())
				Expect(_oriURL).To(Equal(testOrigURL))
				time.Sleep(keyExpire)
				_, err = cacheService.GetCachedURL(testShortenURL)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

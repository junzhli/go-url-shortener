package cache_test

import (
	"fmt"
	rs "github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"time"
	. "url-shortener/internal/cache"
	"url-shortener/internal/config"
)

var _ = Describe("Redis", func() {
	var (
		cache      Redis
		test1Key   string
		test1Value int
		test2Key   string
		test2Value string
		keyExpire  time.Duration
	)

	BeforeEach(func() {
		test1Key = "test1"
		test1Value = 123
		test2Key = "test2"
		test2Value = "456"
		keyExpire = time.Minute

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
	})

	Describe("Set key", func() {
		Context("Set key-pair with value represented by int", func() {
			It("should successfully", func() {
				err := cache.Set(test1Key, test1Value, keyExpire)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Set key-pair with value represented by string", func() {
			It("should successfully", func() {
				err := cache.Set(test2Key, test2Value, keyExpire)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Get key", func() {
		Context("Get key-pair with value represented by int", func() {
			It("should successfully", func() {
				value, err := cache.Get(test1Key)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(strconv.Itoa(test1Value)))
			})
		})

		Context("Set key with value represented by string", func() {
			It("should successfully", func() {
				value, err := cache.Get(test2Key)
				Expect(err).NotTo(HaveOccurred())
				Expect(value).To(Equal(test2Value))
			})
		})
	})

	Describe("Delete key", func() {
		Context("Delete key-pair with value represented by int", func() {
			It("should successfully", func() {
				err := cache.Del(test1Key)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Delete key with value represented by string", func() {
			It("should successfully", func() {
				err := cache.Del(test2Key)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

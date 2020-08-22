package cache

import (
	rs "github.com/go-redis/redis"
	"time"
)

type Service interface {
	GetCachedURL(shortenURL string) (string, error)
	PutCachedURL(shortenURL string, oriURL string, expiration time.Duration) error
	GetCachedURLCount(shortenURL string) (string, error)
	PutCachedURLCount(shortenURL string) (int64, error)
}

var (
	keyCachedUrl      = "KEY_CACHED_URL"
	keyCachedUrlCount = "KEY_CACHED_URL_COUNT"
)

func cachedURLKey(shortenURL string) string {
	return keyCachedUrl + ":" + shortenURL
}

func cachedURLCountKey(shortenURL string) string {
	return keyCachedUrlCount + ":" + shortenURL
}

type service struct {
	redis Redis
}

type NoFoundErr struct {
	detail string
}

func (err *NoFoundErr) Error() string {
	return err.detail + ": Not Found"
}

func (s service) GetCachedURL(shortenURL string) (string, error) {
	oriURL, err := s.redis.Get(cachedURLKey(shortenURL))
	if err != nil {
		if err == rs.Nil {
			return "", &NoFoundErr{detail: "GetCachedURL_" + shortenURL}
		}
		return "", err
	}
	return oriURL, nil
}

func (s service) PutCachedURL(shortenURL string, oriURL string, expiration time.Duration) error {
	return s.redis.Set(cachedURLKey(shortenURL), oriURL, expiration)
}

func (s service) GetCachedURLCount(shortenURL string) (string, error) {
	return s.redis.Get(cachedURLCountKey(shortenURL))
}

func (s service) PutCachedURLCount(shortenURL string) (int64, error) {
	return s.redis.Increment(cachedURLCountKey(shortenURL))
}

func NewService(redis Redis) Service {
	return &service{
		redis: redis,
	}
}

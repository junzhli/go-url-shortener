package cache

import (
	"time"

	rs "github.com/go-redis/redis"
)

// Redis deals with Redis data stuff
type Redis interface {
	Get(key string) (string, error)
	Del(key string) error
	Set(key string, value interface{}, expiration time.Duration) error
	Increment(key string) (int64, error)
	NewTx() rs.Pipeliner
	Ping() error
	Close() error
}

type redis struct {
	client *rs.Client
}

func (r *redis) Get(key string) (string, error) {
	return r.client.Get(key).Result()
}

func (r *redis) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(key, value, expiration).Err()
}

func (r *redis) Del(key string) error {
	return r.client.Del(key).Err()
}

func (r *redis) Close() error {
	return r.client.Close()
}

func (r *redis) NewTx() rs.Pipeliner {
	return r.client.TxPipeline()
}

func (r *redis) Ping() error {
	_, err := r.client.Ping().Result()
	return err
}

func (r *redis) Increment(key string) (int64, error) {
	return r.client.Incr(key).Result()
}

// New creates an instance of Redis
func New(options *rs.Options) Redis {
	return &redis{
		client: rs.NewClient(options),
	}
}

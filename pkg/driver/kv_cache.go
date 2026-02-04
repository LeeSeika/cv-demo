package driver

import (
	"context"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/leeseika/cv-demo/pkg/config"
	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/redis/go-redis/v9"
)

var (
	_kvCacheProvider KVCacheProvider
	_initKVCacheOnce sync.Once
)

type (
	KVCacheProvider interface {
		ProviderType() string
		Get(ctx context.Context, key string) ([]byte, error)
		GetMulti(ctx context.Context, keys []string) (map[string][]byte, error)
		Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
		SetMulti(ctx context.Context, keyValueMap map[string][]byte, expiration time.Duration) error
		SetEmptyValuePlaceholder(ctx context.Context, key string) error
		Delete(ctx context.Context, key string) error
		TTL(ctx context.Context, key string) (time.Duration, error)
	}
)

func InitKVCacheProvider(conf config.Cache) {
	providerType := conf.Provider
	_initKVCacheOnce.Do(func() {
		switch providerType {
		case "redis":
			rdb := redis.NewClient(&redis.Options{
				Addr:     conf.Addr,
				Password: "",
				DB:       0,
			})
			_kvCacheProvider = kvcache.NewRedisProvider(rdb)
		case "memcached":
			mc := memcache.New(conf.Addr)
			_kvCacheProvider = kvcache.NewMemcachedProvider(mc)
		default:
			panic("unknown kv cache provider")
		}
	})
}

func GetKVCacheProvider() KVCacheProvider {
	return _kvCacheProvider
}

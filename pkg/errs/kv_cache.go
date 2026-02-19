package errs

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/leeseika/cv-demo/pkg/driver"
	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/redis/go-redis/v9"
)

func IsKVCacheError(err error, kvCacheError error) bool {
	if errors.Is(err, kvCacheError) {
		return true
	}

	switch driver.GetKVCacheProvider().ProviderType() {
	case "redis":
		return handleRedisError(err, kvCacheError)
	case "memcached":
		return handleMemcachedError(err, kvCacheError)
	}

	return false
}

func handleRedisError(err error, kvCacheError error) bool {
	switch kvCacheError {
	case kvcache.ErrKeyNotFound:
		return errors.Is(err, kvcache.ErrKeyNotFound)
	case kvcache.ErrKeyCacheMissed:
		return errors.Is(err, redis.Nil)
	}

	return errors.Is(err, kvCacheError)
}

func handleMemcachedError(err error, kvCacheError error) bool {
	switch kvCacheError {
	case kvcache.ErrKeyNotFound:
		return errors.Is(err, kvcache.ErrKeyNotFound)
	case kvcache.ErrKeyCacheMissed:
		return errors.Is(err, memcache.ErrCacheMiss)
	}

	return errors.Is(err, kvCacheError)
}

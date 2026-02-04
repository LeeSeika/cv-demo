package errs

import (
	"errors"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/leeseika/cv-demo/pkg/driver"
	kvcache "github.com/leeseika/cv-demo/pkg/driver/kv-cache"
	"github.com/redis/go-redis/v9"
)

func IsKVError(err error, kvError error) bool {
	if errors.Is(err, kvError) {
		return true
	}

	switch driver.GetKVCacheProvider().ProviderType() {
	case "redis":
		return handleRedisError(err, kvError)
	case "memcached":
		return handleMemcachedError(err, kvError)
	}

	return false
}

func handleRedisError(err error, kvError error) bool {
	switch kvError {
	case kvcache.ErrKeyNotFound:
		return errors.Is(err, kvcache.ErrKeyNotFound)
	case kvcache.ErrKeyCacheMissed:
		return errors.Is(err, redis.Nil)
	}

	return errors.Is(err, kvError)
}

func handleMemcachedError(err error, kvError error) bool {
	switch kvError {
	case kvcache.ErrKeyNotFound:
		return errors.Is(err, kvcache.ErrKeyNotFound)
	case kvcache.ErrKeyCacheMissed:
		return errors.Is(err, memcache.ErrCacheMiss)
	}

	return errors.Is(err, kvError)
}

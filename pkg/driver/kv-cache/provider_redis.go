package kvcache

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisProvider struct {
	rdb *redis.Client
}

func NewRedisProvider(rdb *redis.Client) *RedisProvider {
	return &RedisProvider{
		rdb: rdb,
	}
}

func (p *RedisProvider) ProviderType() string {
	return "redis"
}

func (p *RedisProvider) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := p.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare(val, emptyValuePlaceholder) == 1 {
		return nil, ErrKeyNotFound
	}
	return val, nil
}

func (p *RedisProvider) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	if len(keys) == 0 {
		return result, nil
	}
	cmds, err := p.rdb.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	for i, cmd := range cmds {
		if cmd == nil {
			continue
		}
		valStr, ok := cmd.(string)
		if !ok {
			continue
		}
		val := []byte(valStr)
		if subtle.ConstantTimeCompare(val, emptyValuePlaceholder) == 1 {
			continue
		}
		result[keys[i]] = val
	}
	return result, nil
}

func (p *RedisProvider) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if subtle.ConstantTimeCompare(value, emptyValuePlaceholder) == 1 {
		return p.SetEmptyValuePlaceholder(ctx, key)
	}
	return p.rdb.Set(ctx, key, value, expiration).Err()
}

func (p *RedisProvider) SetMulti(ctx context.Context, keyValueMap map[string][]byte, expiration time.Duration) error {
	if len(keyValueMap) == 0 {
		return nil
	}
	pipe := p.rdb.Pipeline()
	for key, value := range keyValueMap {
		pipe.Set(ctx, key, value, expiration)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (p *RedisProvider) SetEmptyValuePlaceholder(ctx context.Context, key string) error {
	return p.rdb.Set(ctx, key, emptyValuePlaceholder, emptyValuePlaceholderTTL).Err()
}

func (p *RedisProvider) Delete(ctx context.Context, key string) error {
	return p.rdb.Del(ctx, key).Err()
}

func (p *RedisProvider) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	return p.rdb.Del(ctx, keys...).Err()
}

func (p *RedisProvider) TTL(ctx context.Context, key string) (time.Duration, error) {
	return p.rdb.TTL(ctx, key).Result()
}

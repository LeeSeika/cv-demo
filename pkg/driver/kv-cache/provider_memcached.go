package kvcache

import (
	"context"
	"crypto/subtle"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcachedProvider struct {
	mc *memcache.Client
}

func (p *MemcachedProvider) ProviderType() string {
	return "memcached"
}

func NewMemcachedProvider(mc *memcache.Client) *MemcachedProvider {
	return &MemcachedProvider{
		mc: mc,
	}
}

func (p *MemcachedProvider) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := p.mc.Get(key)
	if err != nil {
		return nil, err
	}
	if item == nil || len(item.Value) == 0 {
		return nil, ErrKeyCacheMissed
	}
	if subtle.ConstantTimeCompare(item.Value, emptyValuePlaceholder) == 1 {
		return nil, ErrKeyNotFound
	}
	return item.Value, nil
}

func (p *MemcachedProvider) GetMulti(ctx context.Context, keys []string) (map[string][]byte, error) {
	result := make(map[string][]byte)
	if len(keys) == 0 {
		return result, nil
	}
	items, err := p.mc.GetMulti(keys)
	if err != nil {
		return nil, err
	}
	for key, item := range items {
		if item == nil || len(item.Value) == 0 {
			continue
		}
		if subtle.ConstantTimeCompare(item.Value, emptyValuePlaceholder) == 1 {
			continue
		}
		result[key] = item.Value
	}
	return result, nil
}

func (p *MemcachedProvider) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if subtle.ConstantTimeCompare(value, emptyValuePlaceholder) == 1 {
		return p.SetEmptyValuePlaceholder(ctx, key)
	}
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(expiration.Seconds()),
	}
	return p.mc.Set(item)
}

func (p *MemcachedProvider) SetMulti(ctx context.Context, keyValueMap map[string][]byte, expiration time.Duration) error {
	if len(keyValueMap) == 0 {
		return nil
	}
	var firstErr error
	for key, value := range keyValueMap {
		item := &memcache.Item{
			Key:        key,
			Value:      value,
			Expiration: int32(expiration.Seconds()),
		}
		if err := p.mc.Set(item); err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (p *MemcachedProvider) SetEmptyValuePlaceholder(ctx context.Context, key string) error {
	item := &memcache.Item{
		Key:        key,
		Value:      emptyValuePlaceholder,
		Expiration: int32(emptyValuePlaceholderTTL.Seconds()),
	}
	return p.mc.Set(item)
}

func (p *MemcachedProvider) Delete(ctx context.Context, key string) error {
	return p.mc.Delete(key)
}

func (p *MemcachedProvider) TTL(ctx context.Context, key string) (time.Duration, error) {
	item, err := p.mc.Get(key)
	if err != nil {
		return 0, err
	}
	if item == nil {
		return 0, ErrKeyCacheMissed
	}
	if item.Expiration == 0 {
		return 0, nil
	}
	return time.Duration(item.Expiration) * time.Second, nil
}

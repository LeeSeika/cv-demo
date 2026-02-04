package kvcache

import "errors"

type Error struct {
	error
}

func (e Error) Unwrap() error {
	return e.error
}

var (
	// ErrKeyNotFound indicates that the specified key does not exist in the KV store and any other storage layer.
	ErrKeyNotFound = errors.New("key not found")
	// ErrKeyCacheMissed indicates that the specified key does not exist in the KV store, but may exists in other storage layers.
	ErrKeyCacheMissed = errors.New("key cache missed")
)

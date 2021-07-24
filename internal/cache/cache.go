package cache

import (
	"errors"
	"time"

	"github.com/allegro/bigcache/v3"
)

var ErrEntryExpired = errors.New("entry expired")

const timeBinaryRepresentationLen = 15

type Op struct {
	ttl    time.Duration
	hasTTL bool
}

type OpOption func(*Op)

func WithTTL(ttl time.Duration) OpOption {
	return func(op *Op) { op.ttl = ttl }
}

func WithHasTTL() OpOption {
	return func(op *Op) { op.hasTTL = true }
}

type Cache interface {
	Get(key string, ops ...OpOption) ([]byte, error)
	Set(key string, value []byte, ops ...OpOption) error
	Del(key string) error
}

type cache struct {
	cache *bigcache.BigCache
}

func NewCache(cfg bigcache.Config) Cache {
	c, _ := bigcache.NewBigCache(cfg)

	return &cache{cache: c}
}

func (c *cache) Get(key string, ops ...OpOption) ([]byte, error) {
	opt := &Op{}
	for _, o := range ops {
		o(opt)
	}
	if !opt.hasTTL {
		return c.cache.Get(key)
	}

	value, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}
	var bestBefore time.Time
	err = bestBefore.UnmarshalBinary(value[:timeBinaryRepresentationLen])
	if err != nil {
		return nil, err
	}
	if time.Now().After(bestBefore) {
		if err := c.Del(key); err != nil {
			return nil, err
		}
		return nil, ErrEntryExpired
	}
	return value[timeBinaryRepresentationLen:], nil
}

func (c *cache) Set(key string, value []byte, ops ...OpOption) error {
	opt := &Op{}
	for _, o := range ops {
		o(opt)
	}
	if opt.ttl == 0 {
		return c.cache.Set(key, value)
	}

	bestBefore := time.Now().Add(opt.ttl)
	timeBinary, err := bestBefore.MarshalBinary()
	if err != nil {
		return err
	}
	v := append(timeBinary, value...)
	c.cache.Close()
	return c.cache.Set(key, v)
}

func (c *cache) Del(key string) error {
	return c.cache.Delete(key)
}

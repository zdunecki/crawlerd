package cache

import (
	"testing"
	"time"

	"crawlerd/test"
	"github.com/allegro/bigcache/v3"
)

func TestCacheBeforeExpire(t *testing.T) {
	c := NewCache(bigcache.DefaultConfig(10 * time.Minute))

	key := "key"
	value := []byte("value")
	expect := value

	if err := c.Set(key, value, WithTTL(time.Second*2)); err != nil {
		t.Error(err)
	}

	result, err := c.Get("key", WithHasTTL())
	if err != nil {
		t.Error(err)
	}

	test.Diff(t, "value from cache should equal", expect, result)
}

func TestCacheAfterExpire(t *testing.T) {
	c := NewCache(bigcache.DefaultConfig(10 * time.Minute))

	key := "key"
	value := []byte("value")
	var expect []byte

	if err := c.Set(key, value, WithTTL(time.Second*2)); err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second * 2)

	result, err := c.Get("key", WithHasTTL())
	if err != nil && err != ErrEntryExpired {
		t.Error(err)
	}

	test.Diff(t, "value from cache should equal", expect, result)

	if _, err := c.Get("key", WithHasTTL()); err == nil {
		t.Error("key should not exists")
	} else {
		if err == ErrEntryExpired {
			t.Error("entry should return 'not found' error not 'expired'")
		}
	}
}

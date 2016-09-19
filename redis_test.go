package store

import (
	"sync"
	"testing"
	"time"

	"gopkg.in/redis.v4"
)

type testRedisClient struct {
	mu     sync.RWMutex
	values map[string]string
}

func (r *testRedisClient) Get(key string) (string, error) {
	r.mu.RLock()
	v, ok := r.values[key]
	r.mu.RUnlock()
	if !ok {
		return "", redis.Nil
	}
	return v, nil
}

func (r *testRedisClient) Set(key string, value string, expiration time.Duration) error {
	r.mu.Lock()
	r.values[key] = value
	r.mu.Unlock()
	return nil
}

func (r *testRedisClient) Del(key string) (bool, error) {
	r.mu.Lock()
	_, ok := r.values[key]
	delete(r.values, key)
	r.mu.Unlock()
	return ok, nil
}

func TestRedisGetDel(t *testing.T) {
	store := NewRedis(RedisOptions{
		Client: nil,
		Prefix: "entry_",
		Age:    time.Second * 60,
	})

	// for testing
	store.client = &testRedisClient{values: make(map[string]string)}

	for i := 0; i < 3; i++ {
		ret := store.Get("key", func() interface{} {
			return "ok"
		})
		rs := ret.(string)
		if rs != "ok" {
			t.Errorf("got %s expect ok", rs)
		}
	}

	store.Del("key")

	for i := 0; i < 3; i++ {
		ret := store.Get("key", func() interface{} {
			return "ok2"
		})
		rs := ret.(string)
		if rs != "ok2" {
			t.Errorf("got %s expect ok2", rs)
		}
	}
}

package store

import (
	"testing"

	"gopkg.in/redis.v4"
)

type testRedisClient struct {
	store *Memory
}

func (r *testRedisClient) Get(key string) (string, error) {
	v, ok := r.store.Get(key)
	if !ok {
		return "", redis.Nil
	}
	return v.(string), nil
}

func (r *testRedisClient) Set(key string, value string) error {
	r.store.Set(key, value)
	return nil
}

func (r *testRedisClient) Del(keys ...string) (int64, error) {
	for _, key := range keys {
		r.store.Del(key)
	}
	return int64(len(keys)), nil
}

func (r *testRedisClient) Incr(field string, incr int64) (int64, error) {
	return r.store.Incr(field, incr), nil
}

func (r *testRedisClient) IncrF(field string, incr float64) (float64, error) {
	return r.store.IncrF(field, incr), nil
}

func (r *testRedisClient) Clear() error {
	r.store.Clear()
	return nil
}

func TestRedisFetch(t *testing.T) {
	store := NewRedis(nil, "test/")

	// for testing
	store.client = &testRedisClient{store: NewMemory()}

	for i := 0; i < 3; i++ {
		ret, err := store.Fetch("key", func() string {
			return "ok"
		})
		if err != nil {
			t.Error(err)
		} else if ret != "ok" {
			t.Errorf("got %s expect ok", ret)
		}
	}

	store.Del("key")

	for i := 0; i < 3; i++ {
		ret, err := store.Fetch("key", func() string {
			return "ok2"
		})
		if err != nil {
			t.Error(err)
		} else if ret != "ok2" {
			t.Errorf("got %s expect ok2", ret)
		}
	}
}
